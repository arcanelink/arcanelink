# ArcaneLink Web Client

Web client for ArcaneLink distributed instant messaging protocol.

## Features

- User authentication (login/register)
- Direct messaging (P2P)
- Group chat (Rooms)
- HTTP long polling for real-time updates
- Presence status
- Responsive design

## Tech Stack

- React 18
- TypeScript
- Vite
- Zustand (state management)
- React Router

## Getting Started

### Install Dependencies

```bash
npm install
```

### Development

```bash
npm run dev
```

The app will be available at http://localhost:3000

### Build

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## Configuration

The API endpoint is configured in `vite.config.ts`. By default, it proxies to `http://localhost:8080`.

## Project Structure

```
src/
├── api/           # API client and services
├── components/    # React components
├── pages/         # Page components
├── store/         # Zustand stores
├── types/         # TypeScript types
├── utils/         # Utility functions
├── App.tsx        # Main app component
└── main.tsx       # Entry point
```
