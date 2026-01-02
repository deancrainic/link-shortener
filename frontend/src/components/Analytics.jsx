import { useEffect, useState } from 'react'
import './Analytics.css'

const formatCount = (value) => (Number.isFinite(value) ? value : 0)

function Analytics({ formatExpiry, refreshKey }) {
  const [links, setLinks] = useState([])
  const [linksError, setLinksError] = useState('')
  const [linksLoading, setLinksLoading] = useState(false)
  const [lookupCode, setLookupCode] = useState('')
  const [lookupResult, setLookupResult] = useState(null)
  const [lookupError, setLookupError] = useState('')
  const [lookupLoading, setLookupLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('all')
  const [copiedShort, setCopiedShort] = useState(false)
  const [copiedOriginal, setCopiedOriginal] = useState(false)

  const fetchLinks = async () => {
    setLinksLoading(true)
    setLinksError('')
    try {
      const response = await fetch('/api/links')
      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || 'Failed to load links')
      }
      const data = await response.json()
      setLinks(Array.isArray(data) ? data : [])
    } catch (err) {
      setLinksError(err.message || 'Failed to load links.')
    } finally {
      setLinksLoading(false)
    }
  }

  const fetchLinkDetails = async (code) => {
    if (!code) return
    setLookupLoading(true)
    setLookupError('')
    setLookupResult(null)
    try {
      const response = await fetch(`/api/links/${encodeURIComponent(code)}`)
      if (!response.ok) {
        const message = await response.text()
        throw new Error(message || 'Link not found')
      }
      const data = await response.json()
      setLookupResult(data)
    } catch (err) {
      setLookupError(err.message || 'Failed to load link details.')
    } finally {
      setLookupLoading(false)
    }
  }

  useEffect(() => {
    fetchLinks()
  }, [refreshKey])

  return (
    <section className="analytics">
      <div className="analytics-head">
        <div>
          <h2>Analytics</h2>
          <p className="subtle">Switch between the list and a detail view.</p>
        </div>
        <div className="tabs" role="tablist" aria-label="Analytics views">
          <button
            className={`tab ${activeTab === 'all' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveTab('all')}
            role="tab"
            aria-selected={activeTab === 'all'}
          >
            All links
          </button>
          <button
            className={`tab ${activeTab === 'lookup' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveTab('lookup')}
            role="tab"
            aria-selected={activeTab === 'lookup'}
          >
            Lookup
          </button>
        </div>
      </div>

      {activeTab === 'all' ? (
        <>
          <div className="analytics-actions">
            <button
              className="ghost"
              type="button"
              onClick={fetchLinks}
              disabled={linksLoading}
            >
              {linksLoading ? 'Refreshing...' : 'Refresh'}
            </button>
          </div>

          {linksError ? <div className="alert error">{linksError}</div> : null}

          <div className="table">
            <div className="table-row table-head">
              <span>Code</span>
              <span>Original</span>
              <span>Expires</span>
              <span>Total</span>
              <span>Unique</span>
            </div>
            {linksLoading ? (
              <div className="table-row">
                <span className="muted">Loading links...</span>
              </div>
            ) : links.length === 0 ? (
              <div className="table-row">
                <span className="muted">No links yet.</span>
              </div>
            ) : (
              links.map((link) => (
                <div className="table-row" key={link.code}>
                  <span className="mono">{link.code}</span>
                  <span className="truncate">{link.originalUrl}</span>
                  <span>{formatExpiry(link.expiresAt)}</span>
                  <span>{formatCount(link.totalClicks)}</span>
                  <span>{formatCount(link.uniqueVisitors)}</span>
                </div>
              ))
            )}
          </div>
        </>
      ) : (
        <>
          <form
            className="lookup"
            onSubmit={(event) => {
              event.preventDefault()
              fetchLinkDetails(lookupCode.trim())
            }}
          >
            <input
              type="text"
              placeholder="Enter short code"
              value={lookupCode}
              onChange={(event) => setLookupCode(event.target.value)}
            />
            <button className="ghost" type="submit" disabled={lookupLoading}>
              {lookupLoading ? 'Searching...' : 'Search'}
            </button>
          </form>

          {lookupError ? <div className="alert error">{lookupError}</div> : null}

          {lookupResult ? (
            <div className="details">
              <div className="details-main">
                <div className="details-item details-short">
                  <span className="label">Short URL</span>
                  <div className="details-url-block">
                    <strong className="mono">{lookupResult.shortUrl}</strong>
                    <button
                      className="copy"
                      type="button"
                      onClick={async () => {
                        try {
                          await navigator.clipboard?.writeText(
                            lookupResult.shortUrl
                          )
                          setCopiedShort(true)
                          window.setTimeout(() => setCopiedShort(false), 1500)
                        } catch {
                          setCopiedShort(false)
                        }
                      }}
                    >
                      {copiedShort ? 'Copied ✓' : 'Copy'}
                    </button>
                  </div>
                </div>
                <div className="details-item details-original">
                  <span className="label">Original</span>
                  <div className="details-url-block">
                    <strong
                      className="details-url"
                      title={lookupResult.originalUrl}
                    >
                      {lookupResult.originalUrl}
                    </strong>
                    <button
                      className="copy"
                      type="button"
                      onClick={async () => {
                        try {
                          await navigator.clipboard?.writeText(
                            lookupResult.originalUrl
                          )
                          setCopiedOriginal(true)
                          window.setTimeout(
                            () => setCopiedOriginal(false),
                            1500
                          )
                        } catch {
                          setCopiedOriginal(false)
                        }
                      }}
                    >
                      {copiedOriginal ? 'Copied ✓' : 'Copy'}
                    </button>
                  </div>
                </div>
                <div className="details-qr">
                  <span className="label">QR code</span>
                  <div className="qr">
                    <img
                      src={lookupResult.qrCode}
                      alt="QR code for short link"
                    />
                  </div>
                </div>
                <div className="details-item details-meta">
                  <div>
                    <span className="label">Created</span>
                    <strong>{formatExpiry(lookupResult.createdAt)}</strong>
                  </div>
                  <div>
                    <span className="label">Expires</span>
                    <strong>{formatExpiry(lookupResult.expiresAt)}</strong>
                  </div>
                </div>
              </div>
              <div className="details-stats">
                <div>
                  <span className="label">Total clicks</span>
                  <strong>{formatCount(lookupResult.totalClicks)}</strong>
                </div>
                <div>
                  <span className="label">Unique visitors</span>
                  <strong>{formatCount(lookupResult.uniqueVisitors)}</strong>
                </div>
                <div>
                  <span className="label">Last accessed</span>
                  <strong>
                    {lookupResult.lastAccessed
                      ? formatExpiry(lookupResult.lastAccessed)
                      : 'Never'}
                  </strong>
                </div>
              </div>
              <div className="details-countries">
                <span className="label">Countries</span>
                {lookupResult.countryCounts &&
                Object.keys(lookupResult.countryCounts).length > 0 ? (
                  <div className="chips">
                    {Object.entries(lookupResult.countryCounts).map(
                      ([country, count]) => (
                        <span className="chip" key={country}>
                          {country}: {count}
                        </span>
                      )
                    )}
                  </div>
                ) : (
                  <p className="muted">No location data yet.</p>
                )}
              </div>
            </div>
          ) : null}
        </>
      )}
    </section>
  )
}

export default Analytics
