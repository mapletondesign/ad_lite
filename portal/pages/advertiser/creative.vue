<script setup lang="ts">
definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const file = ref<File | null>(null)
const uploading = ref(false)
const uploadedUrl = ref('')
const errorMessage = ref('')

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
  uploadedUrl.value = ''
  errorMessage.value = ''
}

async function handleUpload() {
  if (!file.value) return
  uploading.value = true
  errorMessage.value = ''
  uploadedUrl.value = ''

  const form = new FormData()
  form.append('file', file.value)

  try {
    const result = await apiFetch<{ url: string }>('/api/v1/assets/upload', {
      method: 'POST',
      body: form,
    })
    uploadedUrl.value = result.url
  } catch {
    errorMessage.value = 'Upload failed. Accepted formats: JPEG, PNG, GIF, WebP, MP4, WebM.'
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="h4 fw-bold mb-4">Upload Creative</h1>

    <div class="card" style="max-width: 520px;">
      <div class="card-body">
        <div class="mb-3">
          <label class="form-label" for="creative-file">
            Ad Creative <span class="text-muted small">(image or video)</span>
          </label>
          <input
            id="creative-file"
            class="form-control"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm"
            @change="onFileChange"
          />
          <div class="form-text">Accepted: JPEG, PNG, GIF, WebP, MP4, WebM. Max 100 MB.</div>
        </div>

        <div v-if="errorMessage" class="alert alert-danger py-2 small">{{ errorMessage }}</div>

        <div v-if="uploadedUrl" class="alert alert-success py-2 small">
          Uploaded successfully.
          <a :href="uploadedUrl" target="_blank" class="alert-link ms-1">View file</a>
        </div>

        <button
          class="btn btn-primary"
          type="button"
          :disabled="!file || uploading"
          @click="handleUpload"
        >
          {{ uploading ? 'Uploading…' : 'Upload' }}
        </button>
      </div>
    </div>

    <div v-if="uploadedUrl" class="mt-4">
      <div class="text-muted small mb-2">Preview</div>
      <video
        v-if="file?.type.startsWith('video/')"
        :src="uploadedUrl"
        controls
        class="rounded border"
        style="max-width: 480px; max-height: 270px;"
      />
      <img
        v-else
        :src="uploadedUrl"
        alt="Uploaded creative"
        class="rounded border"
        style="max-width: 480px; max-height: 270px; object-fit: contain;"
      />
    </div>
  </div>
</template>
