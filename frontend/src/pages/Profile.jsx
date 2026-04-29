import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import { apiGet, apiPost, apiPut } from '../api';

export default function Profile() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const [ftpValue, setFtpValue] = useState('');
  const [ftpDate, setFtpDate] = useState(() => new Date().toISOString().split('T')[0]);
  const [ftpSource, setFtpSource] = useState('');
  const [rampPower, setRampPower] = useState('');
  const [rampResult, setRampResult] = useState(null);
  const [ftpHistory, setFtpHistory] = useState([]);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const loadFtpHistory = useCallback(async () => {
    try {
      const data = await apiGet('/api/user/ftp-history');
      setFtpHistory(Array.isArray(data) ? data : []);
    } catch {
      // Silently fail on history load
    }
  }, []);

  useEffect(() => {
    loadFtpHistory();
  }, [loadFtpHistory]);

  async function handleSetFtp(e) {
    e.preventDefault();
    setError('');
    setSuccess('');

    const value = parseInt(ftpValue, 10);
    if (!value || value < 50 || value > 500) {
      setError('FTP must be between 50 and 500 watts.');
      return;
    }

    try {
      await apiPost('/api/user/ftp-history', {
        ftp: value,
        recorded_at: ftpDate,
        source: ftpSource || 'manual entry',
      });
      await apiPut('/api/user/profile', { ftp: value });
      setSuccess(`FTP set to ${value}W`);
      setFtpValue('');
      setFtpSource('');
      loadFtpHistory();
    } catch (err) {
      setError(err.message || 'Failed to save FTP');
    }
  }

  function handleCalculateRamp() {
    setError('');
    setSuccess('');
    const power = parseInt(rampPower, 10);
    if (!power || power < 50 || power > 1000) {
      setError('Max 1-minute power must be between 50 and 1000 watts.');
      return;
    }
    const estimated = Math.round(power * 0.75);
    setRampResult(estimated);
  }

  async function handleConfirmRamp() {
    if (!rampResult) return;
    setError('');
    setSuccess('');

    try {
      await apiPost('/api/user/ftp-from-ramp', {
        max_1min_power: parseInt(rampPower, 10),
      });
      setSuccess(`FTP set to ${rampResult}W (from ramp test)`);
      setRampPower('');
      setRampResult(null);
      loadFtpHistory();
    } catch (err) {
      setError(err.message || 'Failed to save FTP from ramp test');
    }
  }

  function handleLogout() {
    logout();
    navigate('/login');
  }

  return (
    <div className="profile-page">
      <header className="profile-header">
        <h1>User Profile</h1>
        <div className="profile-header-actions">
          <button className="btn btn-secondary" onClick={() => navigate('/editor')}>
            Editor
          </button>
          <button className="btn btn-danger" onClick={handleLogout}>
            Log Out
          </button>
        </div>
      </header>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <section className="profile-section">
        <h2>Account Info</h2>
        <div className="profile-info">
          <p>
            <strong>Name:</strong> {user?.display_name || 'Loading...'}
          </p>
          <p>
            <strong>Email:</strong> {user?.email || 'Loading...'}
          </p>
          <p>
            <strong>Current FTP:</strong> {user?.ftp || '---'}W
          </p>
        </div>
      </section>

      <section className="profile-section">
        <h2>Set FTP Directly</h2>
        <form className="ftp-form" onSubmit={handleSetFtp}>
          <div className="form-group">
            <label htmlFor="ftp-value">FTP (watts)</label>
            <input
              id="ftp-value"
              type="number"
              min="50"
              max="500"
              value={ftpValue}
              onChange={(e) => setFtpValue(e.target.value)}
              placeholder="e.g. 250"
            />
          </div>
          <div className="form-group">
            <label htmlFor="ftp-date">Date</label>
            <input
              id="ftp-date"
              type="date"
              value={ftpDate}
              onChange={(e) => setFtpDate(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label htmlFor="ftp-source">Source (optional)</label>
            <input
              id="ftp-source"
              type="text"
              value={ftpSource}
              onChange={(e) => setFtpSource(e.target.value)}
              placeholder="e.g. 20-min test"
            />
          </div>
          <button type="submit" className="btn btn-primary">
            Save FTP
          </button>
        </form>
      </section>

      <section className="profile-section">
        <h2>Calculate FTP from Ramp Test</h2>
        <div className="ramp-form">
          <div className="form-group">
            <label htmlFor="ramp-power">Max 1-minute Power (watts)</label>
            <input
              id="ramp-power"
              type="number"
              min="50"
              max="1000"
              value={rampPower}
              onChange={(e) => {
                setRampPower(e.target.value);
                setRampResult(null);
              }}
              placeholder="e.g. 350"
            />
          </div>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleCalculateRamp}
          >
            Calculate
          </button>
          {rampResult !== null && (
            <div className="ramp-result">
              <p>
                Estimated FTP: <strong>{rampResult}W</strong> (75% of {rampPower}W)
              </p>
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleConfirmRamp}
              >
                Confirm & Save
              </button>
            </div>
          )}
        </div>
      </section>

      <section className="profile-section">
        <h2>FTP History</h2>
        {ftpHistory.length === 0 ? (
          <p className="empty-state">No FTP history recorded yet.</p>
        ) : (
          <table className="ftp-history-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>FTP (W)</th>
                <th>Source</th>
              </tr>
            </thead>
            <tbody>
              {ftpHistory.map((entry, index) => (
                <tr key={entry.id || index}>
                  <td>{entry.recorded_at}</td>
                  <td>{entry.ftp}</td>
                  <td>{entry.source || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}
