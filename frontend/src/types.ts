export interface User {
  id: number
  name: string
  username: string | null
  email: string
  avatar_path: string | null
  created_at: string
}

export interface Workspace {
  id: number
  name: string
  created_by: number
  created_at: string
  role?: string
}

export interface InviteResponse {
  already_member: boolean
  email_sent: boolean
  invite_link?: string
}

export interface InvitePreview {
  workspace_name: string
  email: string
  already_accepted: boolean
}

export interface WorkspaceMember {
  workspace_id: number
  user_id: number
  name: string
  email: string
  avatar_path: string | null
  role: string
  joined_at: string
}

export interface Board {
  id: number
  workspace_id: number
  name: string
  background_color: string
  created_by: number
  created_at: string
  archived_at?: string | null
}

export interface BoardMember {
  board_id: number
  user_id: number
  name: string
  email: string
  avatar_path: string | null
  role: string
}

export interface Label {
  id: number
  board_id: number
  name: string
  color: string
}

interface CardBase {
  id: number
  list_id: number
  board_id: number
  title: string
  description: string
  position: number
  due_date: string | null
  start_date: string | null
  is_completed: boolean
  cover_color: string | null
  created_by: number
  created_at: string
  updated_at: string
}

export type ArchivedCard = CardBase

export interface CardSummary extends CardBase {
  member_ids: number[]
  label_ids: number[]
  checklist_total: number
  checklist_checked: number
}

export interface ListDetail {
  id: number
  board_id: number
  name: string
  position: number
  cards: CardSummary[]
}

export interface BoardDetail extends Board {
  lists: ListDetail[]
  labels: Label[]
  members: BoardMember[]
}

export interface ChecklistItem {
  id: number
  checklist_id: number
  text: string
  is_checked: boolean
  position: number
}

export interface ChecklistDetail {
  id: number
  card_id: number
  title: string
  position: number
  items: ChecklistItem[]
}

export interface Comment {
  id: number
  card_id: number
  user_id: number
  user_name: string
  body: string
  created_at: string
}

export interface Attachment {
  id: number
  card_id: number
  file_name: string
  file_path: string
  file_size: number
  uploaded_by: number
  uploader_name: string
  created_at: string
}

export interface Activity {
  id: number
  card_id: number
  user_id: number
  user_name: string
  message: string
  created_at: string
}

export interface CardDetail extends CardBase {
  members: BoardMember[]
  labels: Label[]
  checklists: ChecklistDetail[]
  comments: Comment[]
  attachments: Attachment[]
  activities: Activity[]
}
