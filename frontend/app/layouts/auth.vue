<template>
  <div class="min-h-dvh bg-canvas lg:grid lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
    <aside class="relative hidden overflow-hidden bg-ink-950 p-10 text-white lg:flex lg:flex-col lg:justify-between">
      <div aria-hidden="true" class="absolute inset-0 grid-paper-light opacity-60" />
      <div
        aria-hidden="true"
        class="absolute -bottom-40 -left-24 h-[420px] w-[420px] rounded-full bg-brand-500/20 blur-[100px]"
      />

      <div class="relative">
        <NuxtLink to="/" aria-label="based on logistics — home">
          <AppLogo tone="light" :size="38" />
        </NuxtLink>
      </div>

      <div class="relative max-w-md">
        <p class="text-[10.5px] font-semibold tracking-[0.22em] text-brand-200 uppercase">
          Logistics, connected
        </p>
        <h1 class="mt-4 text-[clamp(1.9rem,3vw,2.5rem)] leading-[1.08] font-semibold tracking-[-0.04em]">
          Run your operations. Reach more customers.
        </h1>
        <p class="mt-4 text-[14px] leading-[1.7] text-white/45">
          One workspace for branches, staff, shipments, pricing and marketplace bookings — and one
          marketplace for the customers who need them moved.
        </p>
        <ul class="mt-7 space-y-2.5 text-[13px] text-white/55">
          <li class="flex items-center gap-2.5">
            <span class="size-1.5 rounded-full bg-brand-500" /> Built for Nigerian intercity routes
          </li>
          <li class="flex items-center gap-2.5">
            <span class="size-1.5 rounded-full bg-brand-500" /> Separate company workspaces and permissions
          </li>
          <li class="flex items-center gap-2.5">
            <span class="size-1.5 rounded-full bg-brand-500" /> Marketplace bookings and operations together
          </li>
        </ul>
      </div>

      <div class="relative flex items-center justify-between text-[11.5px] text-white/30">
        <span>© {{ new Date().getFullYear() }} based on logistics</span>
        <NuxtLink to="/" class="transition hover:text-white">Back to site</NuxtLink>
      </div>
    </aside>

    <div class="flex min-h-dvh flex-col">
      <header class="flex items-center justify-between px-5 py-5 lg:hidden">
        <NuxtLink to="/" aria-label="based on logistics — home">
          <AppLogo :size="32" />
        </NuxtLink>
        <div class="flex items-center gap-2">
          <ThemeToggle />
          <NuxtLink to="/" class="text-[12.5px] font-medium text-ink-900/70 hover:text-ink-900">
            Back to site
          </NuxtLink>
        </div>
      </header>

      <main class="flex flex-1 items-center justify-center px-5 py-8 sm:px-8">
        <div class="w-full max-w-[440px]">
          <div class="mb-4 hidden justify-end lg:flex">
            <ThemeToggle />
          </div>
          <slot />
        </div>
      </main>

      <footer class="px-5 pb-6 text-center text-[11px] text-ink-900/35 sm:px-8">
        Secure workspace access · based on logistics (BOL)
      </footer>
    </div>
  </div>
</template>
