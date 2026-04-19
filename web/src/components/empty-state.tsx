interface Props {
  icon?: string
  title: string
  subtitle?: string
}

export function EmptyState({ icon, title, subtitle }: Props) {
  return (
    <div style={{ textAlign: 'center', marginTop: '5rem', color: '#999' }}>
      {icon && <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>{icon}</div>}
      <p style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem', color: '#ddd' }}>{title}</p>
      {subtitle && <p style={{ fontSize: '0.9rem' }}>{subtitle}</p>}
    </div>
  )
}
