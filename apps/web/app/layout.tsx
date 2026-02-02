import type { ReactNode } from 'react';
import './styles.css';

export const metadata = {
  title: 'Recruitment Platform',
  description: 'Candidate marketplace dashboard'
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="app-shell">
          <header className="app-header">
            <div>
              <p className="app-title">Recruitment Platform</p>
              <p className="app-subtitle">AI-powered candidate marketplace</p>
            </div>
            <nav className="app-nav">
              <a href="#candidates">Candidates</a>
              <a href="#search">Search</a>
              <a href="#workflow">Interview Requests</a>
            </nav>
          </header>
          <main>{children}</main>
        </div>
      </body>
    </html>
  );
}
