const noop = () => {}

export const createRiskScanPoller = ({
  requestStatus,
  onProgress = noop,
  onComplete = noop,
  onError = noop,
  interval = 2500,
  timers = globalThis
}) => {
  if (typeof requestStatus !== 'function') throw new TypeError('requestStatus is required')

  let timer
  let generation = 0
  let activeRunId = 0

  const clearTimer = () => {
    if (timer === undefined) return
    timers.clearTimeout(timer)
    timer = undefined
  }

  const stop = () => {
    generation += 1
    activeRunId = 0
    clearTimer()
  }

  const schedule = (currentGeneration) => {
    if (currentGeneration !== generation || !activeRunId) return
    clearTimer()
    timer = timers.setTimeout(() => {
      timer = undefined
      return poll(currentGeneration)
    }, interval)
  }

  const poll = async (currentGeneration) => {
    if (currentGeneration !== generation || !activeRunId) return
    const requestedRunId = activeRunId

    try {
      const run = await requestStatus(requestedRunId)
      if (currentGeneration !== generation || requestedRunId !== activeRunId) return
      if (!run) {
        schedule(currentGeneration)
        return
      }

      await onProgress(run)
      if (currentGeneration !== generation || requestedRunId !== activeRunId) return
      if (run.status === 'running') {
        schedule(currentGeneration)
        return
      }

      activeRunId = 0
      generation += 1
      clearTimer()
      await onComplete(run)
    } catch (error) {
      if (currentGeneration !== generation || requestedRunId !== activeRunId) return
      await onError(error)
      schedule(currentGeneration)
    }
  }

  const start = (runId) => {
    stop()
    const normalizedRunId = Number(runId)
    if (!Number.isInteger(normalizedRunId) || normalizedRunId <= 0) return
    activeRunId = normalizedRunId
    const currentGeneration = ++generation
    schedule(currentGeneration)
  }

  return { start, stop }
}
