interface DistanceBadgeProps {
  distance: number
}

function toRelevancePercent(distance: number): number {
  const clamped = Math.min(1, Math.max(0, distance))
  return Math.round((1 - clamped) * 100)
}

export function DistanceBadge({ distance }: DistanceBadgeProps): React.ReactElement {
  const percent = toRelevancePercent(distance)

  return (
    <span data-testid="distance-badge">
      {percent}%
    </span>
  )
}
