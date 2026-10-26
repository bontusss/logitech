<script setup lang="ts">
useHead({ title: 'Home' })

definePageMeta({ layout: 'default' })

const { user, logout } = useAuth()
const signingOut = ref(false)

async function signOut() {
  signingOut.value = true
  try {
    await logout()
  } finally {
    signingOut.value = false
  }
}

const now = new Date()
const formattedDate = computed(() => {
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  return `${days[now.getDay()]} ${now.getDate()} ${months[now.getMonth()]}`
})

const formattedTime = computed(() => {
  const h = now.getHours().toString().padStart(2, '0')
  const m = now.getMinutes().toString().padStart(2, '0')
  return `${h}:${m}`
})

const userName = computed(() => {
  if (!user.value) return ''
  const parts = user.value.full_name.split(' ')
  return parts[0] || user.value.full_name
})
</script>

<template>
  <div v-if="user" class="min-h-dvh flex flex-col md:flex-row">
    <!-- left sidebar (desktop only) -->
    <aside class="hidden md:flex flex-col border-r border-line bg-surface" style="width: 380px;">
      <div class="flex flex-col gap-6 p-5">
        <!-- user info -->
        <div>
          <p class="text-label text-ink-64">Signed in as</p>
          <h2 class="mt-1 text-h5 text-ink">{{ userName }}</h2>
          <button
            type="button"
            class="mt-2 inline-flex items-center gap-1 text-label text-ink-64 underline decoration-ink underline-offset-4 transition hover:text-blue"
            @click="signOut"
          >
            Sign out
          </button>
        </div>

        <!-- prompt -->
        <div>
          <h3 class="text-h4 text-ink">What do you want to do today?</h3>
        </div>

        <!-- send a parcel -->
        <div class="rounded-lg bg-canvas p-4">
          <h4 class="text-label font-medium text-ink">Send a parcel</h4>
          <p class="mt-1 text-label text-ink-64">For when a customer is bringing goods to send.</p>
          <UiButton
            type="button"
            variant="tertiary"
            class="mt-3"
          >
            Send
          </UiButton>
        </div>

        <!-- receive a parcel -->
        <div class="rounded-lg bg-canvas p-4">
          <h4 class="text-label font-medium text-ink">Receive a parcel</h4>
          <p class="mt-1 text-label text-ink-64">For when goods have arrived with a driver</p>
          <div class="mt-4">
            <UiButton
              type="button"
            >
              Take a photo
              <UiKbd>R</UiKbd>
            </UiButton>
          </div>
          <div class="mt-4">
            <button
              type="button"
              class="inline-flex items-center gap-1 text-label text-ink-64 underline decoration-ink underline-offset-4 transition hover:text-blue"
            >
              Or enter manually
              <UiKbd>M</UiKbd>
            </button>
          </div>
        </div>
      </div>

      <!-- sidebar footer -->
      <div class="mt-auto flex items-center justify-between gap-4 border-t border-line p-5">
        <span class="text-label text-ink-64">Galopin, Aba</span>
        <span class="flex items-center gap-1 text-label text-ink-64">
          {{ formattedDate }}
          <span class="text-ink-40">•</span>
          {{ formattedTime }}
          <svg class="ml-1 size-4 text-ink-64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 6v6l4 2"/>
          </svg>
        </span>
      </div>
    </aside>

    <!-- main content -->
    <main class="flex-1 bg-white/80 p-5 lg:p-8 flex-1 md:ml-[15%] md:pb-20">
      <div class="max-w-2xl">
        <!-- search bar -->
        <UiSearchBar />

        <!-- needs attention -->
        <section class="mt-8">
          <div class="flex items-center gap-2">
            <h2 class="text-h4 text-ink text-[1.125rem]">Needs attention</h2>
            <button class="flex size-6 items-center justify-center rounded-full bg-ink/5 text-label text-ink-64 transition hover:bg-ink/10" aria-label="Help">
              <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <circle cx="12" cy="12" r="10"/>
                <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
                <path d="M12 17h.01"/>
              </svg>
            </button>
          </div>
          <p class="mt-1 text-label text-ink-64">Nothing needs checking.</p>
        </section>

        <!-- pending collection -->
        <section class="mt-8">
          <div class="flex items-center gap-2">
            <h2 class="text-h4 text-ink text-[1.125rem]">Pending collection</h2>
            <button class="flex size-6 items-center justify-center rounded-full bg-ink/5 text-label text-ink-64 transition hover:bg-ink/10" aria-label="Help">
              <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <circle cx="12" cy="12" r="10"/>
                <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
                <path d="M12 17h.01"/>
              </svg>
            </button>
          </div>
          <p class="mt-1 text-label text-ink-64">No goods waiting. New arrivals show here.</p>
        </section>

        <!-- how goods get here -->
        <section class="mt-10">
          <h2 class="text-h4 text-ink">How goods get here</h2>
          <div class="mt-4 flex gap-4">
            <!-- card 01 -->
            <div class="flex w-1/3 flex-col rounded-lg border border-line bg-surface p-4">
              <div class="flex h-20 w-full items-center justify-center rounded-md bg-ink/5">
                <svg class="size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" x2="8" y1="13" y2="13"/>
                  <line x1="16" x2="8" y1="17" y2="17"/>
                </svg>
                <svg class="ml-2 size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <rect x="5" y="2" width="14" height="20" rx="2" ry="2"/>
                  <line x1="12" x2="12" y1="18" y2="18"/>
                </svg>
              </div>
              <p class="mt-2 text-meta font-medium text-blue">01</p>
              <h3 class="mt-1 text-p1 font-medium text-ink">Take a photo of the manifest</h3>
              <p class="mt-1 text-label text-ink-64">Use the camera or your phone.</p>
            </div>

            <!-- card 02 -->
            <div class="flex w-1/3 flex-col rounded-lg border border-line bg-surface p-4">
              <div class="flex h-20 w-full items-center justify-center rounded-md bg-ink/5 relative">
                <svg class="size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" x2="8" y1="13" y2="13"/>
                  <line x1="16" x2="8" y1="17" y2="17"/>
                </svg>
                <span class="absolute right-2 top-2 flex size-5 items-center justify-center rounded-full bg-danger text-white text-xs font-bold">!</span>
              </div>
              <p class="mt-2 text-meta font-medium text-blue">02</p>
              <h3 class="mt-1 text-p1 font-medium text-ink">Check the manifest</h3>
              <p class="mt-1 text-label text-ink-64">Fix any line the photo got wrong.</p>
            </div>

            <!-- card 03 -->
            <div class="flex w-1/3 flex-col rounded-lg border border-line bg-surface p-4">
              <div class="flex h-20 w-full items-center justify-center rounded-md bg-ink/5 relative">
                <svg class="size-8 text-ink-40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <path d="M3 9h18"/>
                  <path d="M9 21V9"/>
                </svg>
                <span class="absolute right-2 top-2 flex size-5 items-center justify-center rounded-full bg-blue text-white text-xs font-bold">✓</span>
                <span class="absolute right-8 top-2 text-meta text-ink-64 text-xs">2 of 3 here</span>
              </div>
              <p class="mt-2 text-meta font-medium text-blue">03</p>
              <h3 class="mt-1 text-p1 font-medium text-ink">Confirm what arrived</h3>
              <p class="mt-1 text-label text-ink-64">Count what arrived. We tell the receivers.</p>
            </div>
          </div>
        </section>

        <div class="mt-8 flex-1"></div>
      </div>
    </main>

    <!-- bottom nav (mobile only) -->
    <nav class="md:hidden fixed bottom-0 left-0 right-0 border-t border-line bg-surface z-10">
      <div class="flex items-center justify-around h-16">
        <!-- profile -->
        <NuxtLink
          to="/profile"
          class="flex flex-col items-center gap-1 py-2"
        >
          <svg class="size-6 text-blue-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
          <span class="text-xs text-ink-64">{{ userName }}</span>
        </NuxtLink>

        <!-- send -->
        <NuxtLink
          to="/send"
          class="flex flex-col items-center gap-1 py-2"
        >
          <svg class="size-6 text-blue-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <line x1="12" x2="12" y1="5" y2="19"/>
            <line x1="5" x2="19" y1="12" y2="12"/>
          </svg>
          <span class="text-xs text-ink-64">Send</span>
        </NuxtLink>

        <!-- receive -->
        <NuxtLink
          to="/receive"
          class="flex flex-col items-center gap-1 py-2"
        >
          <svg class="size-6 text-blue-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" x2="12" y1="15" y2="3"/>
          </svg>
          <span class="text-xs text-ink-64">Receive</span>
        </NuxtLink>

        <!-- settings -->
        <NuxtLink
          to="/settings"
          class="flex flex-col items-center gap-1 py-2"
        >
          <svg class="size-6 text-blue-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9c0 .64-.52 1.16-1.17 1.34l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82V3a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51 1z"/>
          </svg>
          <span class="text-xs text-ink-64">Settings</span>
        </NuxtLink>
      </div>
    </nav>
  </div>
</template>
