import { CalendarDays, CheckSquare } from 'lucide-react'
import Avatar from './Avatar'
import { formatDate } from '../utils/date'
import type { BoardMember, CardSummary, Label } from '../types'

interface Props {
  card: CardSummary
  labels: Label[]
  members: BoardMember[]
}

export default function CardChip({ card, labels, members }: Props) {
  const cardLabels = labels.filter((l) => card.label_ids.includes(l.id))
  const cardMembers = members.filter((m) => card.member_ids.includes(m.user_id))
  const hasChecklist = card.checklist_total > 0
  const checklistDone = hasChecklist && card.checklist_checked === card.checklist_total

  return (
    <div className="bg-white rounded shadow-sm overflow-hidden cursor-pointer hover:shadow-md transition text-sm select-none">
      {card.cover_color && <div className="h-8" style={{ backgroundColor: card.cover_color }} />}
      <div className="px-2.5 py-2">
        {cardLabels.length > 0 && (
          <div className="flex gap-1 mb-1.5 flex-wrap">
            {cardLabels.map((l) => (
              <span key={l.id} className="h-2 w-8 rounded-full" style={{ backgroundColor: l.color }} title={l.name} />
            ))}
          </div>
        )}
        <p className={card.is_completed ? 'text-gray-400 line-through' : 'text-gray-800'}>{card.title}</p>
        <div className="flex items-center justify-between mt-1.5">
          <div className="flex items-center gap-2 text-xs">
            {card.due_date && (
              <span
                className={`flex items-center gap-0.5 rounded px-1 ${
                  card.is_completed ? 'bg-green-100 text-green-700' : 'text-gray-500'
                }`}
              >
                <CalendarDays size={12} />
                {formatDate(card.due_date, { month: 'short', day: 'numeric' })}
              </span>
            )}
            {hasChecklist && (
              <span
                className={`flex items-center gap-0.5 rounded px-1 ${
                  checklistDone ? 'bg-green-100 text-green-700' : 'text-gray-500'
                }`}
              >
                <CheckSquare size={12} />
                {card.checklist_checked}/{card.checklist_total}
              </span>
            )}
          </div>
          {cardMembers.length > 0 && (
            <div className="flex -space-x-1">
              {cardMembers.slice(0, 3).map((m) => (
                <Avatar key={m.user_id} name={m.name} avatarPath={m.avatar_path} size={20} className="border border-white" />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
