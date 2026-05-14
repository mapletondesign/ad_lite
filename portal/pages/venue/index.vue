<script setup lang="ts">
definePageMeta({ middleware: 'role' })
</script>

<template>
  <div class="page">
    <h1 class="page__title">Venue Portal</h1>
    <p class="coming-soon">Coming soon.</p>
  </div>
</template>

<style scoped>
.page {
  max-width: 700px;
}

.page__title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 0.75rem;
}

.coming-soon {
  font-size: 1rem;
  color: var(--color-text-muted);
}
</style>
