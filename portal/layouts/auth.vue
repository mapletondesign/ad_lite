<template>
  <div class="auth-bg">
    <div class="auth-card">
      <div class="auth-brand">
        <span class="auth-brand__name">AdLite</span>
        <span class="auth-brand__tagline">DOOH Advertising Platform</span>
      </div>
      <slot />
    </div>
  </div>
</template>

<style scoped>
.auth-bg {
  min-height: 100vh;
  background-color: var(--color-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}

.auth-card {
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 2.5rem 2rem;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
}

.auth-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 2rem;
  gap: 0.25rem;
}

.auth-brand__name {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-primary);
  letter-spacing: -0.02em;
}

.auth-brand__tagline {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
</style>
