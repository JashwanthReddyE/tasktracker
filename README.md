# Task Tracker 🚀

![Svelte](https://img.shields.io/badge/svelte-%23f1413d.svg?style=for-the-badge&logo=svelte&logoColor=white)
![TailwindCSS](https://imgshields.io/badge/tailwindcss-%2338B2AC.svg?style=for-the-badge&logo=tailwind-css&logoColor=white)
![Supabase](https://img.shields.io/badge/Supabase-3ECF8E?style=for-the-badge&logo=supabase&logoColor=white)

A modern, high-performance, and serverless Kanban board application.

This project was recently migrated from a legacy Go/SQLite architecture to a modern **SvelteKit** and **Supabase** stack, featuring real-time synchronization, native HTML5 drag-and-drop, and a sleek glassmorphism UI.

## ✨ Features

- **Kanban Board**: To Do, In Progress, and Completed columns with drag-and-drop support.
- **Real-Time Sync**: Instant updates across all devices using Supabase Realtime subscriptions.
- **Rich Aesthetics**: Beautiful Tailwind CSS styling with dynamic light and dark modes.
- **Serverless Architecture**: Built on SvelteKit (App Router) and highly optimized for Vercel edge deployments.
- **Custom Authentication**: Secure HttpOnly server-side cookies powered by `@supabase/ssr`.

---

## 🛠️ Tech Stack

- **Frontend**: Svelte 5 (Runes), Tailwind CSS, TypeScript
- **Backend/API**: SvelteKit Server Actions
- **Database**: Supabase PostgreSQL
- **Realtime**: Supabase Channels

---

## 🚀 Getting Started

### 1. Prerequisites
- [Node.js](https://nodejs.org/) (v18 or higher)
- A free [Supabase](https://supabase.com/) account.

### 2. Database Setup
1. Create a new project in Supabase.
2. Navigate to the **SQL Editor** in your Supabase dashboard.
3. Execute the schema migration SQL (provided in the `implementation_plan.md` or as guided during the migration phase) to create the required tables and Row Level Security (RLS) policies.

### 3. Local Development

Clone the repository and install dependencies:
```bash
git clone https://github.com/YOUR_USERNAME/tasktracker.git
cd tasktracker
npm install
```

Set up your environment variables by copying the example file:
```bash
cp .env.example .env
```
Open `.env` and fill in your Supabase credentials:
```env
PUBLIC_SUPABASE_URL="https://your-project-id.supabase.co"
PUBLIC_SUPABASE_ANON_KEY="your-anon-key"
```

Start the development server:
```bash
npm run dev
```

Visit `http://localhost:5173` in your browser.

---

## ☁️ Deployment

This project is optimized for deployment on **Vercel**.

1. Create a [Vercel](https://vercel.com/) account.
2. Import this GitHub repository into Vercel.
3. The Framework Preset should automatically be detected as **SvelteKit**.
4. In the **Environment Variables** section, add your `PUBLIC_SUPABASE_URL` and `PUBLIC_SUPABASE_ANON_KEY`.
5. Click **Deploy**.

---

## 📄 License
MIT License. See [LICENSE](LICENSE) for more information.
