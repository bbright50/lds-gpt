package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ConferenceTalk is one scraped general-conference talk (or sustaining, or
// music number, or any other item linked from a conference landing page).
// Filtering non-talks is left to consumers — preserving everything keeps
// the rawest possible snapshot on disk.
type ConferenceTalk struct {
	URL        string                   `json:"url"`
	Slug       string                   `json:"slug"`                 // last URL segment, e.g. "57nelson"
	Year       int                      `json:"year"`
	Month      int                      `json:"month"`                // 4 or 10
	Session    string                   `json:"session"`              // "Sunday Morning Session"
	Speaker    string                   `json:"speaker"`              // "Russell M. Nelson" (leading "By " / honorifics stripped)
	Role       string                   `json:"role,omitempty"`       // "President of The Church of Jesus Christ of Latter-day Saints"
	Title      string                   `json:"title"`
	Kicker     string                   `json:"kicker,omitempty"`     // optional subtitle / standfirst
	Paragraphs []TalkParagraph          `json:"paragraphs"`
	Footnotes  map[string]TalkFootnote  `json:"footnotes"`            // marker → note (marker is "1", "2", …)
}

// TalkParagraph carries one body paragraph plus any footnote markers
// referenced inline. Paragraph numbers are sequential (p1, p2, …) assigned
// at scrape time — the site's own paragraph IDs are hashed, so we cannot
// cite them stably without re-indexing here.
type TalkParagraph struct {
	ID              string   `json:"id"`
	Text            string   `json:"text"`
	FootnoteMarkers []string `json:"footnote_markers"`
}

// TalkFootnote is the resolved endnote body plus a best-effort extraction
// of scripture references embedded in the prose (for later linking into
// the graph). References stay empty if none are detected.
type TalkFootnote struct {
	Text       string   `json:"text"`
	References []string `json:"references,omitempty"`
}

// TalkRef is the compact record returned by ScrapeConferenceSession — just
// enough to feed ScrapeConferenceTalk and group talks by session.
type TalkRef struct {
	URL     string `json:"url"`
	Slug    string `json:"slug"`
	Session string `json:"session"`
	Speaker string `json:"speaker"` // from landing-page subtitle; confirmed / refined on talk page
	Title   string `json:"title"`
}

// ScrapeConferenceSession fetches a conference landing page
// (https://…/study/general-conference/YYYY/MM?lang=eng) and returns every
// talk reference found, tagged with its session heading. Ordering follows
// the page (sessions top-to-bottom, talks within a session in order).
func ScrapeConferenceSession(ctx context.Context, year, month int, cacheDir string) ([]TalkRef, error) {
	rawURL := fmt.Sprintf("https://www.churchofjesuschrist.org/study/general-conference/%d/%02d?lang=eng", year, month)
	doc, _, err := fetchDocument(ctx, rawURL, cacheDir)
	if err != nil {
		return nil, fmt.Errorf("fetch conference %d-%02d: %w", year, month, err)
	}

	talkPathPrefix := fmt.Sprintf("/study/general-conference/%d/%02d/", year, month)

	var refs []TalkRef
	seen := map[string]bool{}

	// The site groups talks under <li data-content-type="general-conference-session">.
	// Each such <li> starts with a session-header anchor (slug ending in
	// "-session", e.g. saturday-morning-session) followed by per-talk <a>
	// elements. We skip the session-header link and keep the rest.
	doc.Find(`li[data-content-type="general-conference-session"]`).Each(func(_ int, session *goquery.Selection) {
		sessionTitle := normalizeWhitespace(session.Find("p.title").First().Text())

		session.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
			href, _ := a.Attr("href")
			if !strings.HasPrefix(href, talkPathPrefix) {
				return
			}
			slug := extractTalkSlug(href)
			if slug == "" || strings.HasSuffix(slug, "-session") {
				return
			}
			fullURL := "https://www.churchofjesuschrist.org" + href
			if seen[fullURL] {
				return
			}
			seen[fullURL] = true

			// Title sits in the anchor's itemTitle span; speaker in the
			// sibling subtitle. Class hashes (itemTitle-MXhtV,
			// subtitle-LKtQp) may rotate on site rebuilds, so fall back
			// to structural selectors if the hashed names miss.
			title := normalizeWhitespace(a.Find("span").First().Text())
			if title == "" {
				title = normalizeWhitespace(a.Find("p").First().Text())
			}
			speaker := normalizeWhitespace(a.Find("p").Last().Text())
			// If the last <p> happens to be the title (no speaker present),
			// guard against duplicating it into the speaker field.
			if speaker == title {
				speaker = ""
			}

			refs = append(refs, TalkRef{
				URL:     fullURL,
				Slug:    slug,
				Session: sessionTitle,
				Speaker: speaker,
				Title:   title,
			})
		})
	})

	return refs, nil
}

// ScrapeConferenceTalk fetches a single talk page and returns the parsed
// record. session/year/month come from the caller because they live on
// the landing page, not the talk page itself.
func ScrapeConferenceTalk(ctx context.Context, ref TalkRef, year, month int, cacheDir string) (ConferenceTalk, bool, error) {
	doc, cached, err := fetchDocument(ctx, ref.URL, cacheDir)
	if err != nil {
		return ConferenceTalk{}, false, fmt.Errorf("fetch talk %s: %w", ref.URL, err)
	}

	title := extractTalkTitle(doc)
	if title == "" {
		title = ref.Title
	}
	speaker := extractTalkSpeaker(doc)
	if speaker == "" {
		speaker = cleanSpeakerPrefix(ref.Speaker)
	}
	role := normalizeWhitespace(doc.Find("p.author-role").First().Text())
	kicker := normalizeWhitespace(doc.Find("p.kicker").First().Text())

	paragraphs := extractTalkParagraphs(doc)
	refTexts := extractTalkReferenceTexts(doc)
	footnotes := extractTalkFootnotes(doc, refTexts)

	return ConferenceTalk{
		URL:        ref.URL,
		Slug:       ref.Slug,
		Year:       year,
		Month:      month,
		Session:    ref.Session,
		Speaker:    speaker,
		Role:       role,
		Title:      title,
		Kicker:     kicker,
		Paragraphs: paragraphs,
		Footnotes:  footnotes,
	}, cached, nil
}

// --- Selectors ---

func extractTalkTitle(doc *goquery.Document) string {
	// First <h1> inside the article body. The talk's own h1 carries a
	// hashed id (p_xxxx) and no explicit class, so select the first one
	// that isn't inside navigation chrome.
	h1 := doc.Find("article h1").First()
	if h1.Length() == 0 {
		h1 = doc.Find("h1").First()
	}
	return normalizeWhitespace(h1.Text())
}

func extractTalkSpeaker(doc *goquery.Document) string {
	raw := normalizeWhitespace(doc.Find("p.author-name").First().Text())
	return cleanSpeakerPrefix(raw)
}

// cleanSpeakerPrefix strips leading "By " from a byline. It deliberately
// preserves honorifics ("President", "Elder", "Sister", etc.) — the
// downstream graph can split speaker and honorific later if needed;
// dropping them here is lossy.
func cleanSpeakerPrefix(s string) string {
	s = strings.TrimSpace(s)
	const prefix = "By "
	if strings.HasPrefix(s, prefix) {
		return strings.TrimSpace(s[len(prefix):])
	}
	return s
}

// extractTalkParagraphs returns the talk's body in reading order. The
// study site uses <p> elements with hashed ids (p_xxxx) for every piece
// of prose — header, byline, role, kicker, body, footer notes — so we
// scope to the "body-block" container when present and otherwise walk
// the article's <p> children while filtering out byline/kicker/notes.
func extractTalkParagraphs(doc *goquery.Document) []TalkParagraph {
	// Prefer the body-block container — it wraps only the main prose.
	container := doc.Find("div.body-block").First()
	if container.Length() == 0 {
		// Fallback: the whole article minus chrome.
		container = doc.Find("article").First()
	}

	var out []TalkParagraph
	idx := 1
	container.Find("p").Each(func(_ int, p *goquery.Selection) {
		// Skip non-body paragraphs that live inside the same container:
		// byline, role, kicker, note-list, "Notes" title, and anything
		// inside footer.notes.
		if p.HasClass("author-name") || p.HasClass("author-role") ||
			p.HasClass("kicker") || p.HasClass("title") {
			return
		}
		if p.ParentsFiltered("footer.notes").Length() > 0 {
			return
		}

		markers := collectTalkMarkers(p)
		text := buildTalkParagraphText(p)
		if text == "" {
			return
		}

		out = append(out, TalkParagraph{
			ID:              fmt.Sprintf("p%d", idx),
			Text:            text,
			FootnoteMarkers: markers,
		})
		idx++
	})
	return out
}

func buildTalkParagraphText(sel *goquery.Selection) string {
	clone := sel.Clone()
	clone.Find("sup.marker").Remove()
	clone.Find("a.note-ref").Each(func(_ int, a *goquery.Selection) {
		// Keep the anchor's surrounding text flow; markers were already
		// stripped above, and the <a> itself typically has no visible
		// text beyond the sup we just removed.
		a.ReplaceWithSelection(a.Contents())
	})
	return normalizeWhitespace(clone.Text())
}

func collectTalkMarkers(p *goquery.Selection) []string {
	var markers []string
	p.Find("a.note-ref sup.marker").Each(func(_ int, sup *goquery.Selection) {
		if v, ok := sup.Attr("data-value"); ok && v != "" {
			markers = append(markers, v)
		}
	})
	return markers
}

// extractTalkReferenceTexts collects the short in-prose wording
// associated with each footnote marker, keyed by the note id ("note1",
// "note2", …). Mirrors extractReferenceTexts for scriptures but pulls
// the text from the <a class="note-ref"> anchor's readable siblings —
// which on talks is usually empty because the prose context is the
// surrounding sentence. Kept for parity; footnotes without a reference
// text just fall back to "".
func extractTalkReferenceTexts(doc *goquery.Document) map[string]string {
	refs := map[string]string{}
	doc.Find("a.note-ref").Each(func(_ int, a *goquery.Selection) {
		href, ok := a.Attr("href")
		if !ok {
			return
		}
		_, noteID, found := strings.Cut(href, "#")
		if !found {
			return
		}
		clone := a.Clone()
		clone.Find("sup.marker").Remove()
		text := normalizeWhitespace(clone.Text())
		if text != "" {
			refs[noteID] = text
		}
	})
	return refs
}

// extractTalkFootnotes walks footer.notes > ol > li[data-full-marker].
// Marker values on conference talks are plain numbers ("1", "2", …), so
// we strip the trailing "." that sometimes appears in data-full-marker.
func extractTalkFootnotes(doc *goquery.Document, refTexts map[string]string) map[string]TalkFootnote {
	notes := map[string]TalkFootnote{}
	doc.Find("footer.notes li[data-full-marker]").Each(func(_ int, li *goquery.Selection) {
		marker, _ := li.Attr("data-full-marker")
		marker = strings.TrimSuffix(strings.TrimSpace(marker), ".")
		if marker == "" {
			return
		}
		text := normalizeWhitespace(li.Find("p").First().Text())
		if text == "" {
			text = normalizeWhitespace(li.Text())
		}
		fn := TalkFootnote{
			Text:       text,
			References: findScriptureRefsInNote(li),
		}
		_ = refTexts // reserved for future: talks usually have no prose ref-text
		notes[marker] = fn
	})
	return notes
}

// findScriptureRefsInNote mines the note body for scripture citations via
// `<a href="/study/scriptures/…">` anchors — the church site auto-links
// its own references, so this is reliable. Returned values are the
// human-readable anchor text (e.g. "Matthew 5:3").
func findScriptureRefsInNote(li *goquery.Selection) []string {
	seen := map[string]bool{}
	var out []string
	li.Find(`a[href*="/study/scriptures/"]`).Each(func(_ int, a *goquery.Selection) {
		txt := normalizeWhitespace(a.Text())
		if txt == "" || seen[txt] {
			return
		}
		seen[txt] = true
		out = append(out, txt)
	})
	return out
}

// --- Utilities ---

var talkSlugRe = regexp.MustCompile(`/study/general-conference/(\d{4})/(\d{1,2})/([^/?#]+)`)

// extractTalkSlug returns the final path segment of a conference-talk
// URL — the speaker-suffixed identifier like "57nelson". Strips query
// strings and trailing slashes.
func extractTalkSlug(href string) string {
	m := talkSlugRe.FindStringSubmatch(href)
	if len(m) != 4 {
		return ""
	}
	return m[3]
}

// ParseYearMonth turns "2024-10" (or "2024/10") into year, month ints. Any
// out-of-range month that isn't 4 or 10 is kept verbatim — scrape-time
// validation rejects them, so malformed inputs surface clearly there.
func ParseYearMonth(s string) (year, month int, err error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "-")
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected YYYY-MM, got %q", s)
	}
	y, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("year %q: %w", parts[0], err)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("month %q: %w", parts[1], err)
	}
	return y, m, nil
}

// GenerateConferences returns every (year, month) pair in [from, to]
// inclusive, with month ∈ {4, 10}. The two endpoints can use any month;
// they snap onto the nearest valid conference month (e.g. 2024-03 starts
// at 2024-04, 2024-11 ends at 2024-10).
func GenerateConferences(fromY, fromM, toY, toM int) [][2]int {
	months := []int{4, 10}
	var out [][2]int
	for y := fromY; y <= toY; y++ {
		for _, m := range months {
			if y == fromY && m < fromM {
				continue
			}
			if y == toY && m > toM {
				continue
			}
			out = append(out, [2]int{y, m})
		}
	}
	return out
}
