📄 PRD — DevSync (Daily Engineering Reporting System)
1. 📌 Overview

Nama Produk: DevSync
Tipe: Internal Tool / SaaS-ready
Tujuan:

Memudahkan developer (FE, Backend, Mobile) melaporkan progress harian
Memberikan visibilitas real-time ke Tech Lead
Mengidentifikasi blocker dan bottleneck lebih cepat
2. 👥 User Roles
1. Developer
Submit daily report
Melihat history report sendiri
2. Tech Lead
Melihat semua report
Monitoring blocker & progress
Melihat analytics
3. Admin (opsional)
Manage user & team
3. 🧱 Tech Stack
Backend
Go
REST API (atau gRPC internal)
Frontend
React
Tailwind (UI)
Database
MySQL
Cache / Queue (opsional)
Redis
4. 🎯 Core Features
4.1 Daily Report Submission
Requirement

Developer submit report harian dengan struktur:

Tasks done
Tasks next
Blockers
Acceptance Criteria
1 user hanya bisa submit 1 report per hari
Bisa edit di hari yang sama
Field wajib:
tasks_done
tasks_next
4.2 Report Dashboard (Tech Lead)
Features
List semua report
Filter:
by date
by user
by team
Highlight
User tidak submit report
User dengan blocker
4.3 Blocker Detection
Logic
Jika field blockers tidak kosong → tandai sebagai “⚠️ blocker”
Output
Muncul di dashboard
Bisa difilter
4.4 Notification System (MVP: basic)
Reminder submit report (jam tertentu)
Alert ke lead jika ada blocker
4.5 Report History
Developer bisa lihat history
Pagination
5. 🗄️ Database Design (MySQL)
Table: users
id (PK)
name
email
role (developer, lead, admin)
created_at
Table: reports
id (PK)
user_id (FK)
date (DATE)
tasks_done (TEXT)
tasks_next (TEXT)
blockers (TEXT)
created_at
updated_at

Constraint:

UNIQUE (user_id, date)
Table: teams (optional)
id
name
Table: user_teams (optional)
user_id
team_id
6. 🔌 API Design

Base URL:

/api/v1
6.1 Auth
POST /auth/login
{
  "email": "",
  "password": ""
}

Response:

{
  "token": "jwt_token"
}
6.2 Submit Report
POST /reports
{
  "date": "2026-04-20",
  "tasks_done": "implement login",
  "tasks_next": "implement dashboard",
  "blockers": "API belum ready"
}
6.3 Get My Reports
GET /reports/me
6.4 Get All Reports (Lead)
GET /reports

Query:

?date=2026-04-20&user_id=1
6.5 Update Report
PUT /reports/:id
7. 🏗️ Backend Architecture (Go)

Struktur sederhana:

/cmd
/internal
  /handler
  /service
  /repository
  /model
/pkg
Layering
Handler → HTTP layer
Service → business logic
Repository → DB access
8. 🔄 Flow System
User login
Submit report
Backend:
validate
simpan ke MySQL
(opsional) emit event ke Redis
Dashboard update
9. 🖥️ Frontend (React)
Pages
1. Login Page
2. Developer Dashboard
Form submit report
History list
3. Tech Lead Dashboard
Table report
Filter
Highlight blocker
10. 📊 Future Enhancements
1. Analytics
productivity trend
team performance
2. Integration
Jira
GitHub (auto detect commit)
3. Real-time update
WebSocket
11. ⚠️ Non-Functional Requirements
Performance
API response < 200ms
Security
JWT auth
input validation
Reliability
daily backup DB
12. 🚀 MVP Scope (2–3 minggu)

Wajib:

auth
submit report
dashboard lead
blocker highlight

Tidak wajib:

analytics
integration
real-time