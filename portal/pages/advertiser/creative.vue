<script setup lang="ts">
definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const file        = ref<File | null>(null)
const uploading   = ref(false)
const uploadedUrl = ref('')
const error       = ref('')
const snackbar    = ref(false)

const ACCEPTED = 'image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm'

function onFileChange(val: File | null) {
  file.value = val
  uploadedUrl.value = ''
  error.value = ''
}

async function handleUpload() {
  if (!file.value) return
  uploading.value = true
  error.value = ''
  uploadedUrl.value = ''

  const form = new FormData()
  form.append('file', file.value)

  try {
    const result = await apiFetch<{ url: string }>('/api/v1/assets/upload', {
      method: 'POST',
      body: form,
    })
    uploadedUrl.value = result.url
    snackbar.value = true
  } catch {
    error.value = 'Upload failed. Accepted formats: JPEG, PNG, GIF, WebP, MP4, WebM.'
  } finally {
    uploading.value = false
  }
}

const isVideo = computed(() => file.value?.type.startsWith('video/') ?? false)
</script>

<template>
  <div>
    <div class="text-h5 font-weight-bold mb-6">Upload Creative</div>

    <v-card rounded="lg" style="max-width: 560px;">
      <v-card-text>
        <v-file-input
          v-model="file"
          label="Ad Creative"
          hint="Accepted: JPEG, PNG, GIF, WebP, MP4, WebM. Max 100 MB."
          persistent-hint
          :accept="ACCEPTED"
          variant="outlined"
          density="comfortable"
          prepend-icon=""
          prepend-inner-icon="mdi-paperclip"
          class="mb-2"
          @update:model-value="onFileChange"
        />

        <v-alert v-if="error" type="error" density="compact" class="mb-4" rounded="lg">{{ error }}</v-alert>

        <div class="d-flex justify-end">
          <v-btn
            color="primary"
            :loading="uploading"
            :disabled="!file"
            @click="handleUpload"
          >
            Upload
          </v-btn>
        </div>
      </v-card-text>
    </v-card>

    <div v-if="uploadedUrl" class="mt-6">
      <div class="text-caption text-medium-emphasis mb-2">Preview</div>
      <video
        v-if="isVideo"
        :src="uploadedUrl"
        controls
<<<<<<< HEAD
        class="rounded-lg"
        style="max-width: 480px; max-height: 270px; display: block;"
      />
      <v-img
        v-else
        :src="uploadedUrl"
        alt="Uploaded creative"
<<<<<<< HEAD
        max-width="480"
        max-height="270"
        cover
        rounded="lg"
      />
    </div>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">Creative uploaded successfully.</v-snackbar>
  </div>
</template>
