package dataloader

// BookInfo maps a book abbreviation to its volume and directory slug.
type BookInfo struct {
	Volume string // e.g. "ot", "nt", "bofm", "dc-testament", "pgp"
	Slug   string // e.g. "gen", "1-ne", "dc"
}

// buildAbbreviationMap returns a map from standard abbreviation to BookInfo.
// Each key is the canonical abbreviation used in footnote text (e.g. "Gen.", "1 Ne.", "D&C").
func buildAbbreviationMap() map[string]BookInfo {
	return map[string]BookInfo{
		// Old Testament
		"Gen.":    {Volume: "ot", Slug: "gen"},
		"Ex.":     {Volume: "ot", Slug: "ex"},
		"Lev.":    {Volume: "ot", Slug: "lev"},
		"Num.":    {Volume: "ot", Slug: "num"},
		"Deut.":   {Volume: "ot", Slug: "deut"},
		"Josh.":   {Volume: "ot", Slug: "josh"},
		"Judg.":   {Volume: "ot", Slug: "judg"},
		"Ruth":    {Volume: "ot", Slug: "ruth"},
		"1 Sam.":  {Volume: "ot", Slug: "1-sam"},
		"2 Sam.":  {Volume: "ot", Slug: "2-sam"},
		"1 Kgs.":  {Volume: "ot", Slug: "1-kgs"},
		"2 Kgs.":  {Volume: "ot", Slug: "2-kgs"},
		"1 Chr.":  {Volume: "ot", Slug: "1-chr"},
		"2 Chr.":  {Volume: "ot", Slug: "2-chr"},
		"Ezra":    {Volume: "ot", Slug: "ezra"},
		"Neh.":    {Volume: "ot", Slug: "neh"},
		"Esth.":   {Volume: "ot", Slug: "esth"},
		"Job":     {Volume: "ot", Slug: "job"},
		"Ps.":     {Volume: "ot", Slug: "ps"},
		"Prov.":   {Volume: "ot", Slug: "prov"},
		"Eccl.":   {Volume: "ot", Slug: "eccl"},
		"Song":    {Volume: "ot", Slug: "song"},
		"Isa.":    {Volume: "ot", Slug: "isa"},
		"Jer.":    {Volume: "ot", Slug: "jer"},
		"Lam.":    {Volume: "ot", Slug: "lam"},
		"Ezek.":   {Volume: "ot", Slug: "ezek"},
		"Dan.":    {Volume: "ot", Slug: "dan"},
		"Hosea":   {Volume: "ot", Slug: "hosea"},
		"Joel":    {Volume: "ot", Slug: "joel"},
		"Amos":    {Volume: "ot", Slug: "amos"},
		"Obad.":   {Volume: "ot", Slug: "obad"},
		"Jonah":   {Volume: "ot", Slug: "jonah"},
		"Micah":   {Volume: "ot", Slug: "micah"},
		"Nahum":   {Volume: "ot", Slug: "nahum"},
		"Hab.":    {Volume: "ot", Slug: "hab"},
		"Zeph.":   {Volume: "ot", Slug: "zeph"},
		"Hag.":    {Volume: "ot", Slug: "hag"},
		"Zech.":   {Volume: "ot", Slug: "zech"},
		"Mal.":    {Volume: "ot", Slug: "mal"},

		// New Testament
		"Matt.":   {Volume: "nt", Slug: "matt"},
		"Mark":    {Volume: "nt", Slug: "mark"},
		"Luke":    {Volume: "nt", Slug: "luke"},
		"John":    {Volume: "nt", Slug: "john"},
		"Acts":    {Volume: "nt", Slug: "acts"},
		"Rom.":    {Volume: "nt", Slug: "rom"},
		"1 Cor.":  {Volume: "nt", Slug: "1-cor"},
		"2 Cor.":  {Volume: "nt", Slug: "2-cor"},
		"Gal.":    {Volume: "nt", Slug: "gal"},
		"Eph.":    {Volume: "nt", Slug: "eph"},
		"Philip.": {Volume: "nt", Slug: "philip"},
		"Col.":    {Volume: "nt", Slug: "col"},
		"1 Thes.": {Volume: "nt", Slug: "1-thes"},
		"2 Thes.": {Volume: "nt", Slug: "2-thes"},
		"1 Tim.":  {Volume: "nt", Slug: "1-tim"},
		"2 Tim.":  {Volume: "nt", Slug: "2-tim"},
		"Titus":   {Volume: "nt", Slug: "titus"},
		"Philem.": {Volume: "nt", Slug: "philem"},
		"Heb.":    {Volume: "nt", Slug: "heb"},
		"James":   {Volume: "nt", Slug: "james"},
		"1 Pet.":  {Volume: "nt", Slug: "1-pet"},
		"2 Pet.":  {Volume: "nt", Slug: "2-pet"},
		"1 Jn.":   {Volume: "nt", Slug: "1-jn"},
		"2 Jn.":   {Volume: "nt", Slug: "2-jn"},
		"3 Jn.":   {Volume: "nt", Slug: "3-jn"},
		"Jude":    {Volume: "nt", Slug: "jude"},
		"Rev.":    {Volume: "nt", Slug: "rev"},

		// Book of Mormon
		"1 Ne.":   {Volume: "bofm", Slug: "1-ne"},
		"2 Ne.":   {Volume: "bofm", Slug: "2-ne"},
		"Jacob":   {Volume: "bofm", Slug: "jacob"},
		"Enos":    {Volume: "bofm", Slug: "enos"},
		"Jarom":   {Volume: "bofm", Slug: "jarom"},
		"Omni":    {Volume: "bofm", Slug: "omni"},
		"W of M":  {Volume: "bofm", Slug: "w-of-m"},
		"Mosiah":  {Volume: "bofm", Slug: "mosiah"},
		"Alma":    {Volume: "bofm", Slug: "alma"},
		"Hel.":    {Volume: "bofm", Slug: "hel"},
		"3 Ne.":   {Volume: "bofm", Slug: "3-ne"},
		"4 Ne.":   {Volume: "bofm", Slug: "4-ne"},
		"Morm.":   {Volume: "bofm", Slug: "morm"},
		"Ether":   {Volume: "bofm", Slug: "ether"},
		"Moro.":   {Volume: "bofm", Slug: "moro"},

		// Doctrine and Covenants
		"D&C":     {Volume: "dc-testament", Slug: "dc"},
		"OD":      {Volume: "dc-testament", Slug: "od"},

		// Pearl of Great Price
		"Moses":   {Volume: "pgp", Slug: "moses"},
		"Abr.":    {Volume: "pgp", Slug: "abr"},
		"JS—M":    {Volume: "pgp", Slug: "js-m"},
		"JS—H":    {Volume: "pgp", Slug: "js-h"},
		"A of F":  {Volume: "pgp", Slug: "a-of-f"},
	}
}

// buildSlugToAbbrevMap returns a map from "{volume}/{slug}" to the standard abbreviation.
func buildSlugToAbbrevMap() map[string]string {
	abbrevs := buildAbbreviationMap()
	result := make(map[string]string, len(abbrevs))
	for abbrev, info := range abbrevs {
		key := info.Volume + "/" + info.Slug
		result[key] = abbrev
	}
	return result
}

// volumeDisplayNames maps volume abbreviation to display name.
var volumeDisplayNames = map[string]string{
	"ot":           "Old Testament",
	"nt":           "New Testament",
	"bofm":         "Book of Mormon",
	"dc-testament": "Doctrine and Covenants",
	"pgp":          "Pearl of Great Price",
}

// volumeAbbreviations lists all volume abbreviations in canonical order.
var volumeAbbreviations = []string{"ot", "nt", "bofm", "dc-testament", "pgp"}

// buildBookDisplayNameMap returns a map from JSON "book" display name to volume/slug.
// This handles ambiguous names from the scraped data (e.g. "Kings" could be 1 or 2 Kings).
// Only used for books whose JSON "book" field is unreliable or empty.
func buildBookDisplayNameMap() map[string]BookInfo {
	return map[string]BookInfo{
		// OT
		"Genesis":         {Volume: "ot", Slug: "gen"},
		"Exodus":          {Volume: "ot", Slug: "ex"},
		"Leviticus":       {Volume: "ot", Slug: "lev"},
		"Numbers":         {Volume: "ot", Slug: "num"},
		"Deuteronomy":     {Volume: "ot", Slug: "deut"},
		"Joshua":          {Volume: "ot", Slug: "josh"},
		"Judges":          {Volume: "ot", Slug: "judg"},
		"Ruth":            {Volume: "ot", Slug: "ruth"},
		"1 Samuel":        {Volume: "ot", Slug: "1-sam"},
		"2 Samuel":        {Volume: "ot", Slug: "2-sam"},
		"1 Kings":         {Volume: "ot", Slug: "1-kgs"},
		"2 Kings":         {Volume: "ot", Slug: "2-kgs"},
		"1 Chronicles":    {Volume: "ot", Slug: "1-chr"},
		"2 Chronicles":    {Volume: "ot", Slug: "2-chr"},
		"Ezra":            {Volume: "ot", Slug: "ezra"},
		"Nehemiah":        {Volume: "ot", Slug: "neh"},
		"Esther":          {Volume: "ot", Slug: "esth"},
		"Job":             {Volume: "ot", Slug: "job"},
		"Psalms":          {Volume: "ot", Slug: "ps"},
		"Psalm":           {Volume: "ot", Slug: "ps"},
		"Proverbs":        {Volume: "ot", Slug: "prov"},
		"Ecclesiastes":    {Volume: "ot", Slug: "eccl"},
		"Song of Solomon": {Volume: "ot", Slug: "song"},
		"Isaiah":          {Volume: "ot", Slug: "isa"},
		"Jeremiah":        {Volume: "ot", Slug: "jer"},
		"Lamentations":    {Volume: "ot", Slug: "lam"},
		"Ezekiel":         {Volume: "ot", Slug: "ezek"},
		"Daniel":          {Volume: "ot", Slug: "dan"},
		"Hosea":           {Volume: "ot", Slug: "hosea"},
		"Joel":            {Volume: "ot", Slug: "joel"},
		"Amos":            {Volume: "ot", Slug: "amos"},
		"Obadiah":         {Volume: "ot", Slug: "obad"},
		"Jonah":           {Volume: "ot", Slug: "jonah"},
		"Micah":           {Volume: "ot", Slug: "micah"},
		"Nahum":           {Volume: "ot", Slug: "nahum"},
		"Habakkuk":        {Volume: "ot", Slug: "hab"},
		"Zephaniah":       {Volume: "ot", Slug: "zeph"},
		"Haggai":          {Volume: "ot", Slug: "hag"},
		"Zechariah":       {Volume: "ot", Slug: "zech"},
		"Malachi":         {Volume: "ot", Slug: "mal"},

		// NT
		"St Matthew":        {Volume: "nt", Slug: "matt"},
		"Matthew":           {Volume: "nt", Slug: "matt"},
		"St Mark":           {Volume: "nt", Slug: "mark"},
		"Mark":              {Volume: "nt", Slug: "mark"},
		"St Luke":           {Volume: "nt", Slug: "luke"},
		"Luke":              {Volume: "nt", Slug: "luke"},
		"St John":           {Volume: "nt", Slug: "john"},
		"John":              {Volume: "nt", Slug: "john"},
		"Acts":              {Volume: "nt", Slug: "acts"},
		"Romans":            {Volume: "nt", Slug: "rom"},
		"1 Corinthians":     {Volume: "nt", Slug: "1-cor"},
		"2 Corinthians":     {Volume: "nt", Slug: "2-cor"},
		"Galatians":         {Volume: "nt", Slug: "gal"},
		"Ephesians":         {Volume: "nt", Slug: "eph"},
		"Philippians":       {Volume: "nt", Slug: "philip"},
		"Colossians":        {Volume: "nt", Slug: "col"},
		"1 Thessalonians":   {Volume: "nt", Slug: "1-thes"},
		"2 Thessalonians":   {Volume: "nt", Slug: "2-thes"},
		"1 Timothy":         {Volume: "nt", Slug: "1-tim"},
		"2 Timothy":         {Volume: "nt", Slug: "2-tim"},
		"Titus":             {Volume: "nt", Slug: "titus"},
		"Philemon":          {Volume: "nt", Slug: "philem"},
		"Hebrews":           {Volume: "nt", Slug: "heb"},
		"James":             {Volume: "nt", Slug: "james"},
		"1 Peter":           {Volume: "nt", Slug: "1-pet"},
		"2 Peter":           {Volume: "nt", Slug: "2-pet"},
		"1 John":            {Volume: "nt", Slug: "1-jn"},
		"2 John":            {Volume: "nt", Slug: "2-jn"},
		"3 John":            {Volume: "nt", Slug: "3-jn"},
		"Jude":              {Volume: "nt", Slug: "jude"},
		"Revelation":        {Volume: "nt", Slug: "rev"},

		// Book of Mormon
		"1 Nephi":                          {Volume: "bofm", Slug: "1-ne"},
		"The First Book of Nephi":          {Volume: "bofm", Slug: "1-ne"},
		"2 Nephi":                          {Volume: "bofm", Slug: "2-ne"},
		"The Second Book of Nephi":         {Volume: "bofm", Slug: "2-ne"},
		"Jacob":                            {Volume: "bofm", Slug: "jacob"},
		"The Book of Jacob":                {Volume: "bofm", Slug: "jacob"},
		"Enos":                             {Volume: "bofm", Slug: "enos"},
		"The Book of Enos":                 {Volume: "bofm", Slug: "enos"},
		"Jarom":                            {Volume: "bofm", Slug: "jarom"},
		"The Book of Jarom":                {Volume: "bofm", Slug: "jarom"},
		"Omni":                             {Volume: "bofm", Slug: "omni"},
		"The Book of Omni":                 {Volume: "bofm", Slug: "omni"},
		"Words of Mormon":                  {Volume: "bofm", Slug: "w-of-m"},
		"The Words of Mormon":              {Volume: "bofm", Slug: "w-of-m"},
		"Mosiah":                           {Volume: "bofm", Slug: "mosiah"},
		"The Book of Mosiah":               {Volume: "bofm", Slug: "mosiah"},
		"Alma":                             {Volume: "bofm", Slug: "alma"},
		"The Book of Alma":                 {Volume: "bofm", Slug: "alma"},
		"Helaman":                          {Volume: "bofm", Slug: "hel"},
		"The Book of Helaman":              {Volume: "bofm", Slug: "hel"},
		"3 Nephi":                          {Volume: "bofm", Slug: "3-ne"},
		"Third Nephi":                      {Volume: "bofm", Slug: "3-ne"},
		"4 Nephi":                          {Volume: "bofm", Slug: "4-ne"},
		"Fourth Nephi":                     {Volume: "bofm", Slug: "4-ne"},
		"Mormon":                           {Volume: "bofm", Slug: "morm"},
		"The Book of Mormon":               {Volume: "bofm", Slug: "morm"},
		"Ether":                            {Volume: "bofm", Slug: "ether"},
		"The Book of Ether":                {Volume: "bofm", Slug: "ether"},
		"Moroni":                           {Volume: "bofm", Slug: "moro"},
		"The Book of Moroni":               {Volume: "bofm", Slug: "moro"},

		// D&C
		"Doctrine and Covenants":           {Volume: "dc-testament", Slug: "dc"},
		"Official Declaration":             {Volume: "dc-testament", Slug: "od"},

		// PGP
		"Moses":                            {Volume: "pgp", Slug: "moses"},
		"The Book of Moses":                {Volume: "pgp", Slug: "moses"},
		"Abraham":                          {Volume: "pgp", Slug: "abr"},
		"The Book of Abraham":              {Volume: "pgp", Slug: "abr"},
		"Joseph Smith—Matthew":             {Volume: "pgp", Slug: "js-m"},
		"Joseph Smith—History":             {Volume: "pgp", Slug: "js-h"},
		"Articles of Faith":                {Volume: "pgp", Slug: "a-of-f"},
		"The Articles of Faith":            {Volume: "pgp", Slug: "a-of-f"},
	}
}

// buildJSTBookToSlugMap returns a map from JST "book" field to volume/slug.
// The JST JSON uses full book names like "1 Samuel", "Matthew", etc.
func buildJSTBookToSlugMap() map[string]BookInfo {
	return buildBookDisplayNameMap()
}
