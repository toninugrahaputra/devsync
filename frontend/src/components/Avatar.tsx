import { fileUrl } from '../api/client'

interface Props {
  name: string
  avatarPath?: string | null
  size?: number
  className?: string
}

export default function Avatar({ name, avatarPath, size = 24, className = '' }: Props) {
  const style = { width: size, height: size, fontSize: Math.max(9, size * 0.42) }

  if (avatarPath) {
    return (
      <img
        src={fileUrl(avatarPath)}
        alt={name}
        title={name}
        style={style}
        className={`rounded-full object-cover shrink-0 ${className}`}
      />
    )
  }

  return (
    <span
      title={name}
      style={style}
      className={`rounded-full bg-[#D6AE32] text-gray-900 flex items-center justify-center shrink-0 ${className}`}
    >
      {name.slice(0, 1).toUpperCase()}
    </span>
  )
}
