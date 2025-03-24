# Agent Flow Designer

A Next.js application for designing and connecting components using a visual flow editor.

## Tech Stack

### Frontend
- Next.js with TypeScript
- React
- ReactFlow for flow visualization and manipulation
- shadcn/ui for UI components
- Lucide icons

### Backend
- Go services
- RESTful API

## Getting Started

### Prerequisites

- Node.js (v18+)
- npm
- Go (v1.21+) - for backend services

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/agenticflows.git
cd agenticflows
```

2. Install frontend dependencies
```bash
npm install
```

3. Start the development server
```bash
npm run dev
```

4. Set up the Go backend (requires Go to be installed)
```bash
cd backend
go mod tidy
go run api/main.go
```

## Development

### Frontend

The application is structured as follows:
- `src/components/FlowEditor.tsx` - Main component for ReactFlow integration
- `src/services/api.ts` - API service for interacting with the backend
- `src/app/page.tsx` - Main page layout

### Backend

The Go backend provides a simple API for managing flows:
- GET `/api/flows` - Get all flows
- POST `/api/flows` - Create a new flow
- GET `/api/flows/:id` - Get a specific flow
- PUT `/api/flows/:id` - Update a flow
- DELETE `/api/flows/:id` - Delete a flow

## Features

- Visual flow editor for connecting components
- Saving and loading flow configurations
- RESTful API for managing flows

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
