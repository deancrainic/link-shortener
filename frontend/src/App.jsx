import { useState } from 'react'
import './App.css'
import Analytics from './components/Analytics'

const toRFC3339LocalMidnight = (dateValue) => {
  if (!dateValue) return null
  const [year, month, day] = dateValue.split('-').map(Number)
  if (!year || !month || !day) return null
  const localDate = new Date(year, month - 1, day, 0, 0, 0)
  const offsetMinutes = -localDate.getTimezoneOffset()
  const sign = offsetMinutes >= 0 ? '+' : '-'
  const absOffset = Math.abs(offsetMinutes)
  const offsetHours = String(Math.floor(absOffset / 60)).padStart(2, '0')
  const offsetMins = String(absOffset % 60).padStart(2, '0')
  const yyyy = localDate.getFullYear()
  const mm = String(localDate.getMonth() + 1).padStart(2, '0')
  const dd = String(localDate.getDate()).padStart(2, '0')
  const hh = String(localDate.getHours()).padStart(2, '0')
  const mi = String(localDate.getMinutes()).padStart(2, '0')
  const ss = String(localDate.getSeconds()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}:${ss}${sign}${offsetHours}:${offsetMins}`
}

const buildPayload = (url, customAlias, expiresAt) => {
  const trimmedAlias = customAlias.trim()
  const trimmedExpiry = expiresAt.trim()
  return {
    url: url.trim(),
    customAlias: trimmedAlias,
    expiresAt: trimmedExpiry ? toRFC3339LocalMidnight(trimmedExpiry) : null,
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
  const [copiedResult, setCopiedResult] = useState(false)
  const [analyticsRefresh, setAnalyticsRefresh] = useState(0)

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
      setAnalyticsRefresh((value) => value + 1)
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
                type="date"
                value={expiresAt}
                onChange={(event) => setExpiresAt(event.target.value)}
              />
              <p className="helper">Date only. Leave empty for the default 30 days.</p>
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
              <div className="result-url-row">
                <a
                  href={result.shortUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  {result.shortUrl}
                </a>
                <button
                  className="copy"
                  type="button"
                  onClick={async () => {
                    try {
                      await navigator.clipboard?.writeText(result.shortUrl)
                      setCopiedResult(true)
                      window.setTimeout(() => setCopiedResult(false), 1500)
                    } catch {
                      setCopiedResult(false)
                    }
                  }}
                >
                  {copiedResult ? 'Copied ✓' : 'Copy'}
                </button>
              </div>
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

        <Analytics formatExpiry={formatExpiry} refreshKey={analyticsRefresh} />
      </main>
    </div>
  )
}

export default App
