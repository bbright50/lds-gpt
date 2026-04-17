import type { EntityType } from '../types/search.ts'

interface EntityIconProps {
  entityType: EntityType
}

const ICON_MAP: Record<EntityType, string> = {
  verse_group: '\u2261',
  chapter: '\uD83D\uDCD6',
  topical_guide: '\uD83C\uDFF7\uFE0F',
  bible_dict: '\uD83D\uDCD3',
  index: '\uD83D\uDCCB',
  jst_passage: '\u270F\uFE0F',
  verse: '\uD83D\uDCDC',
}

export function EntityIcon({ entityType }: EntityIconProps): React.ReactElement {
  return (
    <span data-testid="entity-icon" aria-label={entityType}>
      {ICON_MAP[entityType]}
    </span>
  )
}
