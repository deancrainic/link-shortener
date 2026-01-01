import { useState } from 'react'
import './App.css'

const buildPayload = (url, customAlias, expiresAt) => {
  const trimmedAlias = customAlias.trim()
  const trimmedExpiry = expiresAt.trim()
  return {
    url: url.trim(),
    customAlias: trimmedAlias,
    expiresAt: trimmedExpiry ? new Date(trimmedExpiry).toISOString() : null,
  }
}

const formatExpiry = (value) => {
  if (!value) return 'Not set'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function App() {
  const [url, setUrl] = useState('')
  const [customAlias, setCustomAlias] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [result, setResult] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (event) => {
    event.preventDefault()
    setError('')
    setResult(null)
    setLoading(true)

    try {
      const payload = buildPayload(url, customAlias, expiresAt)
      const response = await fetch('/api/shorten', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || 'Failed to shorten link')
      }

      const data = await response.json()
      setResult(data)
    } catch (err) {
      setError(err.message || 'Something went wrong.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <header className="hero">
        <div>
          <p className="eyebrow">LinkShortener</p>
          <h1>URL Shortener, Branded Short Links & Analytics</h1>
          <p className="subhead">
            Create a custom alias, set an expiration date, and grab a QR code
            in one <span className='subhead-url'><a href='https://go.dev/'>GO</a></span>.
          </p>
        </div>
      </header>

      <main className="panel">
        <form className="form" onSubmit={handleSubmit}>
          <div className="field">
            <label htmlFor="url">Destination URL</label>
            <input
              id="url"
              name="url"
              type="url"
              placeholder="https://example.com/your/very/long/link"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              required
            />
            <p className="helper">Must start with http or https.</p>
          </div>

          <div className="grid">
            <div className="field">
              <label htmlFor="alias">Custom alias</label>
              <input
                id="alias"
                name="alias"
                type="text"
                placeholder="my-campaign"
                value={customAlias}
                onChange={(event) => setCustomAlias(event.target.value)}
              />
              <p className="helper">3-30 chars: letters, numbers, _ or -</p>
            </div>

            <div className="field">
              <label htmlFor="expiresAt">Expiration date</label>
              <input
                id="expiresAt"
                name="expiresAt"
                type="datetime-local"
                value={expiresAt}
                onChange={(event) => setExpiresAt(event.target.value)}
              />
              <p className="helper">Leave empty for the default 30 days.</p>
            </div>
          </div>

          <button className="primary" type="submit" disabled={loading}>
            {loading ? 'Shortening...' : 'Shorten'}
          </button>
        </form>

        {error ? <div className="alert error">{error}</div> : null}

        {result ? (
          <section className="result">
            <div>
              <h2>Your short link is ready</h2>
              <p className="result-url">
                <a
                  href={result.shortUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  {result.shortUrl}
                </a>
              </p>
              <div className="result-meta">
                <div>
                  <span>Code</span>
                  <strong>{result.code}</strong>
                </div>
                <div>
                  <span>Expires</span>
                  <strong>{formatExpiry(result.expiresAt)}</strong>
                </div>
              </div>
              <p className="result-original">
                Original: <span>{result.originalUrl}</span>
              </p>
            </div>
            {result.qrCode ? (
              <div className="qr">
                <img src={result.qrCode} alt="QR code for short link" />
              </div>
            ) : null}
          </section>
        ) : null}
      </main>
    </div>
  )
}

export default App
