'use client';

import { useEffect, useState } from 'react';

type Candidate = {
  id: string;
  name: string;
  skills: string[];
  readiness_status: string;
  updated_at: string;
};

type SearchResult = {
  candidate: Candidate;
  score: number;
};

type InterviewRequest = {
  id: string;
  recruiter_id: string;
  candidate_id: string;
  status: string;
  expires_at: string;
};

const candidateApi = process.env.NEXT_PUBLIC_CANDIDATE_API ?? 'http://localhost:8082';
const searchApi = process.env.NEXT_PUBLIC_SEARCH_API ?? 'http://localhost:8084';
const workflowApi = process.env.NEXT_PUBLIC_WORKFLOW_API ?? 'http://localhost:8085';

export default function HomePage() {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [requests, setRequests] = useState<InterviewRequest[]>([]);
  const [candidateForm, setCandidateForm] = useState({ name: '', skills: '', readiness_status: 'verified' });
  const [searchForm, setSearchForm] = useState({ skills: '', readiness_status: 'verified', minimum_score: 1 });
  const [requestForm, setRequestForm] = useState({ recruiter_id: 'recruiter-1', candidate_id: '', expires_in_days: 7 });
  const [statusMessage, setStatusMessage] = useState<string>('');

  useEffect(() => {
    fetchCandidates().catch(() => {
      setStatusMessage('Unable to load candidates. Ensure backend services are running.');
    });
  }, []);

  const fetchCandidates = async () => {
    const response = await fetch(`${candidateApi}/candidates`);
    const data = (await response.json()) as Candidate[];
    setCandidates(data);
  };

  const handleCandidateSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setStatusMessage('');

    const payload = {
      name: candidateForm.name,
      skills: candidateForm.skills.split(',').map((skill) => skill.trim()).filter(Boolean),
      readiness_status: candidateForm.readiness_status
    };

    await fetch(`${candidateApi}/candidates`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    await fetchCandidates();
    setCandidateForm({ name: '', skills: '', readiness_status: candidateForm.readiness_status });
  };

  const handleSearch = async (event: React.FormEvent) => {
    event.preventDefault();
    setStatusMessage('');

    const payload = {
      skills: searchForm.skills.split(',').map((skill) => skill.trim()).filter(Boolean),
      readiness_status: searchForm.readiness_status,
      minimum_score: Number(searchForm.minimum_score)
    };

    const response = await fetch(`${searchApi}/search`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    const data = (await response.json()) as SearchResult[];
    setSearchResults(data);
  };

  const handleRequestSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setStatusMessage('');

    const response = await fetch(`${workflowApi}/requests`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        recruiter_id: requestForm.recruiter_id,
        candidate_id: requestForm.candidate_id,
        expires_in_days: Number(requestForm.expires_in_days)
      })
    });

    const data = (await response.json()) as InterviewRequest;
    setRequests((prev) => [data, ...prev]);
    setRequestForm((prev) => ({ ...prev, candidate_id: '' }));
  };

  return (
    <div className="dashboard">
      <section id="candidates" className="card">
        <h2>Candidate Intake</h2>
        <form onSubmit={handleCandidateSubmit}>
          <label>
            Candidate name
            <input
              value={candidateForm.name}
              onChange={(event) => setCandidateForm({ ...candidateForm, name: event.target.value })}
              placeholder="Jane Doe"
              required
            />
          </label>
          <label>
            Skills (comma-separated)
            <input
              value={candidateForm.skills}
              onChange={(event) => setCandidateForm({ ...candidateForm, skills: event.target.value })}
              placeholder="Go, Kafka, Kubernetes"
            />
          </label>
          <label>
            Readiness
            <select
              value={candidateForm.readiness_status}
              onChange={(event) => setCandidateForm({ ...candidateForm, readiness_status: event.target.value })}
            >
              <option value="verified">Interview-ready</option>
              <option value="unverified">Not interview-ready</option>
            </select>
          </label>
          <button type="submit">Save candidate</button>
        </form>
        <p className="footer-note">Profiles are stored via the candidate-profile service.</p>
      </section>

      <section className="card">
        <h2>Candidate Directory</h2>
        <ul className="data-list">
          {candidates.map((candidate) => (
            <li key={candidate.id}>
              <strong>{candidate.name}</strong>
              <span className="badge">{candidate.readiness_status}</span>
              <div>{candidate.skills.join(', ') || 'No skills listed'}</div>
            </li>
          ))}
        </ul>
      </section>

      <section id="search" className="card">
        <h2>Recruiter Search</h2>
        <form onSubmit={handleSearch}>
          <label>
            Skills
            <input
              value={searchForm.skills}
              onChange={(event) => setSearchForm({ ...searchForm, skills: event.target.value })}
              placeholder="Kafka, Go"
            />
          </label>
          <label>
            Readiness
            <select
              value={searchForm.readiness_status}
              onChange={(event) => setSearchForm({ ...searchForm, readiness_status: event.target.value })}
            >
              <option value="verified">Interview-ready</option>
              <option value="unverified">Not interview-ready</option>
            </select>
          </label>
          <label>
            Minimum score
            <input
              type="number"
              value={searchForm.minimum_score}
              onChange={(event) => setSearchForm({ ...searchForm, minimum_score: Number(event.target.value) })}
              min={0}
              max={10}
            />
          </label>
          <button className="secondary" type="submit">Run search</button>
        </form>
        <ul className="data-list">
          {searchResults.map((result) => (
            <li key={result.candidate.id}>
              <strong>{result.candidate.name}</strong>
              <span className="badge">Score {result.score}</span>
              <div>{result.candidate.skills.join(', ')}</div>
            </li>
          ))}
        </ul>
      </section>

      <section id="workflow" className="card">
        <h2>Interview Requests</h2>
        <form onSubmit={handleRequestSubmit}>
          <label>
            Recruiter ID
            <input
              value={requestForm.recruiter_id}
              onChange={(event) => setRequestForm({ ...requestForm, recruiter_id: event.target.value })}
              required
            />
          </label>
          <label>
            Candidate ID
            <input
              value={requestForm.candidate_id}
              onChange={(event) => setRequestForm({ ...requestForm, candidate_id: event.target.value })}
              placeholder="cand-123"
              required
            />
          </label>
          <label>
            Expiry (days)
            <input
              type="number"
              value={requestForm.expires_in_days}
              onChange={(event) => setRequestForm({ ...requestForm, expires_in_days: Number(event.target.value) })}
              min={1}
              max={30}
            />
          </label>
          <button type="submit">Request interview</button>
        </form>
        <ul className="data-list">
          {requests.map((request) => (
            <li key={request.id}>
              <strong>{request.candidate_id}</strong>
              <span className="badge">{request.status}</span>
              <div>Expires: {new Date(request.expires_at).toLocaleDateString()}</div>
            </li>
          ))}
        </ul>
        <p className="footer-note">Requests are created via recruiter-workflow service.</p>
      </section>

      {statusMessage ? <section className="card">{statusMessage}</section> : null}
    </div>
  );
}
