<script setup lang="ts">
/**
 * Public landing page for based on logistics (BOL).
 * No session, no API calls: everything on this page is static marketing copy
 * plus illustrative demo data in the preview components.
 */
definePageMeta({ layout: 'landing' })

// Absolute URL for the share banner: link previews do not resolve relative paths.
const shareImage = `${useRequestURL().origin}/og-image.png`

useHead({
  title: 'based on logistics (BOL) — the logistics operating system',
  // app.vue appends "· Galopin"; the landing page owns its own title.
  titleTemplate: (title) => title ?? '',
  meta: [
    {
      name: 'description',
      content:
        'based on logistics (BOL) is the operating system for courier, haulage and consolidation businesses — one record for every waybill, bag, dispatch, hand-over and naira across every branch.',
    },
    { property: 'og:type', content: 'website' },
    { property: 'og:site_name', content: 'based on logistics (BOL)' },
    { property: 'og:image', content: shareImage },
    { property: 'og:image:type', content: 'image/png' },
    { property: 'og:image:width', content: '1200' },
    { property: 'og:image:height', content: '630' },
    {
      property: 'og:image:alt',
      content:
        'based on logistics (BOL) — Every parcel. Every movement. Accounted for. The logistics operating system.',
    },
    { property: 'og:title', content: 'Every parcel. Every movement. Accounted for.' },
    {
      property: 'og:description',
      content:
        'Record waybills, dispatch bags and containers, follow custody and reconcile collections across every branch — on one logistics operating system.',
    },
    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:image', content: shareImage },
    { name: 'twitter:title', content: 'Every parcel. Every movement. Accounted for.' },
    {
      name: 'twitter:description',
      content:
        'The logistics operating system for courier, haulage and consolidation businesses. Record, dispatch, receive, reconcile.',
    },
  ],
})

const steps = [
  {
    n: '01',
    title: 'Record at the counter',
    body: 'Staff capture the waybill, sender, receiver and items, take payment or leave a balance, and hand the customer a receipt.',
  },
  {
    n: '02',
    title: 'Build the bag',
    body: 'Group everything heading the same way into a bag or container, with a manifest that lists every parcel inside it.',
  },
  {
    n: '03',
    title: 'Dispatch the movement',
    body: 'Assign the vehicle and driver. The bag leaves — whole or in part — and every parcel inside inherits the leg.',
  },
  {
    n: '04',
    title: 'Receive and reconcile',
    body: 'The receiving branch confirms hand-over. Collections, balances and exceptions roll up into one report.',
  },
]

const modules = [
  {
    icon: 'package',
    title: 'Waybills & records',
    body: 'One record per parcel: sender, receiver, items, declared value and every movement since drop-off.',
  },
  {
    icon: 'truck',
    title: 'Bags & containers',
    body: 'Group parcels into manifests that can be sealed, scanned and dispatched as a single unit.',
  },
  {
    icon: 'filter',
    title: 'Partial dispatch',
    body: 'Hold back the parcels that are not ready without breaking the manifest or the custody trail.',
  },
  {
    icon: 'building',
    title: 'Branch network',
    body: 'Every counter sees what is coming to it, what has arrived and what is waiting for collection.',
  },
  {
    icon: 'check',
    title: 'Custody & hand-over',
    body: 'Who held what, at which branch, and when — with the signature captured at collection.',
  },
  {
    icon: 'wallet',
    title: 'Collections & balances',
    body: 'Part payments, outstanding balances and receipts that reconcile back to the waybill they came from.',
  },
  {
    icon: 'alert',
    title: 'Exceptions & follow-up',
    body: 'Log a missing item, a damaged parcel or a delay, and follow it through to a resolution.',
  },
  {
    icon: 'users',
    title: 'Fleet, drivers & audit',
    body: 'Vehicles, drivers and an activity log of every action taken inside the workspace.',
  },
]

const audiences = [
  {
    title: 'Courier & parcel operators',
    body: 'Counters that take walk-in parcels and hand them to a rider or a line haul the same day.',
  },
  {
    title: 'Haulage & trucking',
    body: 'Loads that leave as one vehicle but belong to many customers with different balances.',
  },
  {
    title: 'Consolidators & forwarders',
    body: 'Bags and containers built over days, dispatched in parts, and reconciled on arrival.',
  },
  {
    title: 'Branch networks & franchises',
    body: 'Many counters, one standard: the same record, the same receipt, the same report.',
  },
]

const tenants = [
  { name: 'Swift Line Logistics', initials: 'SL', branches: '6 branches', tone: 'bg-[#2674da]' },
  { name: 'Kano Express', initials: 'KE', branches: '3 branches', tone: 'bg-[#0f9d8e]' },
  { name: 'Apapa Cargo Co.', initials: 'AC', branches: '9 branches', tone: 'bg-[#7b5cf0]' },
]
</script>

<template>
  <div>
    <!-- ── Hero ─────────────────────────────────────────────────────────── -->
    <section id="top" class="relative overflow-hidden">
      <div
        class="pointer-events-none absolute inset-x-0 top-0 h-[560px] bg-[radial-gradient(760px_420px_at_50%_-8%,rgba(38,116,218,0.10),transparent_70%)]"
        aria-hidden="true"
      />
      <div
        class="landing-grid pointer-events-none absolute inset-x-0 top-0 h-[520px] [mask-image:linear-gradient(to_bottom,black,transparent)]"
        aria-hidden="true"
      />

      <div class="relative mx-auto max-w-[1200px] px-5 pt-16 pb-10 sm:px-8 sm:pt-24">
        <div class="landing-reveal mx-auto max-w-3xl text-center">
          <p
            class="inline-flex items-center gap-2 rounded-full border border-[#0b1424]/10 bg-white/80 px-3 py-1 text-[11.5px] font-medium tracking-[0.02em] text-[#0b1424]/60 backdrop-blur"
          >
            <span class="relative flex size-1.5">
              <span class="absolute inline-flex size-full animate-ping rounded-full bg-[#0f9d8e] opacity-60" />
              <span class="relative inline-flex size-1.5 rounded-full bg-[#0f9d8e]" />
            </span>
            The logistics operating system
          </p>

          <h1
            class="mt-6 text-[clamp(2.2rem,5.6vw,4.1rem)] leading-[1.03] font-semibold tracking-[-0.045em] text-[#0b1424] text-balance"
          >
            Every parcel. Every movement.
            <span
              class="bg-[linear-gradient(100deg,#1b5cab,#2674da_45%,#0f9d8e)] bg-clip-text text-transparent"
            >
              Accounted for.
            </span>
          </h1>

          <p
            class="mx-auto mt-6 max-w-2xl text-[17px] leading-[1.65] text-[#0b1424]/60 text-pretty sm:text-[18px]"
          >
            BOL is the system courier, haulage and consolidation businesses run on: waybills, bags and
            containers, custody between branches, and the money that comes with all of it — in one record.
          </p>

          <div class="mt-8 flex flex-col items-center justify-center gap-2.5 sm:flex-row">
            <LandingActionLink to="/sign-in" size="lg" class="w-full sm:w-auto">
              Get started free
              <UiIcon name="arrowRight" :size="16" />
            </LandingActionLink>
            <LandingActionLink to="#how-it-works" variant="secondary" size="lg" class="w-full sm:w-auto">
              See how it works
            </LandingActionLink>
          </div>

          <p class="mt-4 text-[12.5px] text-[#0b1424]/40">
            Free for one branch · No card required · Works on the phones your staff already carry
          </p>
        </div>

        <div class="landing-fade mt-12 sm:mt-16" style="animation-delay: 160ms">
          <LandingHeroPreview />
        </div>

        <div class="mt-8 flex flex-wrap items-center justify-center gap-x-7 gap-y-3 text-[12.5px] text-[#0b1424]/45">
          <span v-for="item in ['Drop-off to hand-over', 'Bags & containers', 'Partial dispatch', 'Branch reconciliation']" :key="item" class="inline-flex items-center gap-1.5">
            <UiIcon name="check" :size="13" class="text-[#2674da]" />
            {{ item }}
          </span>
        </div>
      </div>
    </section>

    <!-- ── The problem ──────────────────────────────────────────────────── -->
    <section class="border-y border-[#0b1424]/8 bg-[#f7f8fa]">
      <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
        <div class="grid gap-10 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)] lg:gap-16">
          <LandingSectionHeading
            eyebrow="The gap"
            title="A logistics business is not hard to run. It is hard to keep track of."
            lede="Waybills in a book. Dispatch updates in a WhatsApp group. Balances in someone's head. The moment a parcel changes hands, the truth splits into three versions — and the end of the month is spent putting them back together."
          />

          <div class="space-y-3">
            <div
              v-for="item in [
                {
                  title: 'The counter and the road disagree',
                  body: 'A parcel is marked “dispatched” at the branch and “not on board” at the truck. Nobody can say which record is real.',
                },
                {
                  title: 'Money hides in the gaps',
                  body: 'Part payments, balances collected on delivery and receipts that never make it back to the office books.',
                },
                {
                  title: 'Nothing can be proved after the fact',
                  body: 'When a customer calls about a missing item, the answer depends on who remembers handing it over.',
                },
              ]"
              :key="item.title"
              class="rounded-xl border border-[#0b1424]/8 bg-white p-5"
            >
              <h3 class="text-[15px] font-medium tracking-[-0.01em] text-[#0b1424]">{{ item.title }}</h3>
              <p class="mt-1.5 text-[13.5px] leading-[1.65] text-[#0b1424]/55">{{ item.body }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ── How it works ─────────────────────────────────────────────────── -->
    <section id="how-it-works" class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-28">
      <LandingSectionHeading
        eyebrow="How it works"
        title="Four steps, from the counter to the customer."
        lede="The same record moves with the parcel. Each hand-over adds an event instead of starting a new document."
      />

      <ol class="mt-12 grid gap-px overflow-hidden rounded-2xl border border-[#0b1424]/10 bg-[#0b1424]/10 sm:grid-cols-2 lg:grid-cols-4">
        <li v-for="step in steps" :key="step.n" class="bg-white p-6">
          <span class="text-[11px] font-semibold tracking-[0.14em] text-[#2674da] tabular-nums">{{ step.n }}</span>
          <h3 class="mt-4 text-[15.5px] font-medium tracking-[-0.01em] text-[#0b1424]">{{ step.title }}</h3>
          <p class="mt-2 text-[13.5px] leading-[1.65] text-[#0b1424]/55">{{ step.body }}</p>
        </li>
      </ol>
    </section>

    <!-- ── Partial dispatch ─────────────────────────────────────────────── -->
    <section id="partial-dispatch" class="border-y border-[#0b1424]/8 bg-[#f7f9fc]">
      <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-28">
        <div class="max-w-[52rem]">
          <p class="text-[11px] font-semibold tracking-[0.22em] text-[#2674da] uppercase">
            Where most systems fall apart
          </p>
          <h2
            class="mt-3 text-[clamp(1.7rem,3.6vw,2.75rem)] leading-[1.08] font-semibold tracking-[-0.035em] text-[#0b1424] text-balance"
          >
            One bag. Many parcels. One clean dispatch.
          </h2>
          <p class="mt-5 max-w-2xl text-[16.5px] leading-[1.65] text-[#0b1424]/60 text-pretty">
            The truck is loaded, but three parcels are not ready. Most systems force a choice: hold the whole
            bag, or lose the parcels that stayed behind. BOL dispatches what is on board and keeps every held
            parcel exactly where it is — with its own status, its own reason and its own place in the record.
          </p>
        </div>

        <div class="mt-10">
          <LandingDispatchVisual />
        </div>

        <div class="mt-8 grid gap-4 sm:grid-cols-3">
          <div v-for="item in [
            { k: 'Whole or part', v: 'A bag can leave with 6 of its 8 parcels and stay a bag.' },
            { k: 'Nothing collapses', v: 'Held parcels keep their status, balance and branch.' },
            { k: 'Provable later', v: 'The manifest is the record of what actually moved.' },
          ]" :key="item.k" class="rounded-xl border border-[#0b1424]/8 bg-white p-4">
            <p class="text-[12.5px] font-semibold text-[#0b1424]">{{ item.k }}</p>
            <p class="mt-1 text-[12.5px] leading-[1.6] text-[#0b1424]/50">{{ item.v }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ── Modules ──────────────────────────────────────────────────────── -->
    <section id="platform" class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-28">
      <LandingSectionHeading
        eyebrow="The platform"
        title="One system, every part of the operation."
        lede="Not a tracker bolted onto a spreadsheet. Each module writes to the same record, so a parcel never has two versions of the truth."
      />

      <div class="mt-12 grid gap-px overflow-hidden rounded-2xl border border-[#0b1424]/10 bg-[#0b1424]/10 sm:grid-cols-2 lg:grid-cols-4">
        <div v-for="module in modules" :key="module.title" class="bg-white p-5">
          <span class="flex size-9 items-center justify-center rounded-lg bg-[#2674da]/[0.08] text-[#2674da]">
            <UiIcon :name="module.icon" :size="17" />
          </span>
          <h3 class="mt-4 text-[14.5px] font-medium tracking-[-0.01em] text-[#0b1424]">{{ module.title }}</h3>
          <p class="mt-1.5 text-[13px] leading-[1.6] text-[#0b1424]/55">{{ module.body }}</p>
        </div>
      </div>
    </section>

    <!-- ── Audience ─────────────────────────────────────────────────────── -->
    <section id="solutions" class="border-y border-[#0b1424]/8 bg-[#f7f8fa]">
      <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-24">
        <LandingSectionHeading
          eyebrow="Who it is for"
          title="Built for the people who move things."
          lede="If a parcel changes hands more than once before it reaches a customer, this is the system for you."
          align="center"
        />

        <div class="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="audience in audiences" :key="audience.title" class="rounded-xl border border-[#0b1424]/8 bg-white p-5">
            <h3 class="text-[14.5px] font-medium tracking-[-0.01em] text-[#0b1424]">{{ audience.title }}</h3>
            <p class="mt-2 text-[13px] leading-[1.65] text-[#0b1424]/55">{{ audience.body }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ── Multi-tenant / white-label ───────────────────────────────────── -->
    <section id="multi-tenant" class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-28">
      <div class="grid items-center gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(0,0.95fr)] lg:gap-16">
        <div>
          <LandingSectionHeading
            eyebrow="Multi-tenant by design"
            title="Your brand on the front. One engine behind it."
            lede="BOL runs each business as its own workspace. Your logo, your domain, your branches, your staff roles, your price book — isolated from every other tenant, on infrastructure you do not have to maintain."
          />

          <ul class="mt-8 space-y-3">
            <li v-for="line in [
              'Separate workspace, separate data — down to the branch.',
              'Your logo and domain on the screens your staff and customers see.',
              'Roles and permissions set per business, not per user.',
              'Plans and price books configured for how you actually sell.',
            ]" :key="line" class="flex gap-3">
              <span class="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full bg-[#2674da]/10 text-[#2674da]">
                <UiIcon name="check" :size="11" />
              </span>
              <span class="text-[14px] leading-[1.6] text-[#0b1424]/65">{{ line }}</span>
            </li>
          </ul>

          <div class="mt-8">
            <LandingActionLink to="#pricing" variant="secondary">
              See the Business plan
              <UiIcon name="arrowRight" :size="15" />
            </LandingActionLink>
          </div>
        </div>

        <!-- Workspace switcher mock -->
        <div class="rounded-2xl border border-[#0b1424]/10 bg-white p-4 shadow-[0_30px_70px_-45px_rgba(11,20,36,0.45)] sm:p-5">
          <div class="flex items-center justify-between">
            <p class="text-[11px] font-semibold tracking-[0.12em] text-[#0b1424]/40 uppercase">Your workspaces</p>
            <span class="rounded-[4px] border border-[#0b1424]/10 px-1.5 py-0.5 text-[9px] font-medium text-[#0b1424]/40">
              BOL engine
            </span>
          </div>

          <ul class="mt-4 space-y-2">
            <li
              v-for="(tenant, index) in tenants"
              :key="tenant.name"
              class="flex items-center gap-3 rounded-xl border p-3"
              :class="index === 0 ? 'border-[#0b1424]/15 bg-[#fafbfc]' : 'border-[#0b1424]/8'"
            >
              <span
                class="flex size-8 shrink-0 items-center justify-center rounded-lg text-[11px] font-semibold text-white"
                :class="tenant.tone"
              >
                {{ tenant.initials }}
              </span>
              <div class="min-w-0 flex-1">
                <p class="truncate text-[12.5px] font-medium text-[#0b1424]">{{ tenant.name }}</p>
                <p class="truncate text-[10.5px] text-[#0b1424]/40">{{ tenant.branches }} · own pricing</p>
              </div>
              <UiIcon v-if="index === 0" name="check" :size="14" class="shrink-0 text-[#2674da]" />
              <UiIcon v-else name="chevronRight" :size="14" class="shrink-0 text-[#0b1424]/25" />
            </li>
          </ul>

          <div class="mt-4 rounded-xl bg-[#f7f8fa] p-3">
            <p class="text-[10.5px] leading-[1.55] text-[#0b1424]/45">
              Each workspace keeps its own waybills, staff and balances. Switching does not merge a single
              record.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- ── Pricing ──────────────────────────────────────────────────────── -->
    <section id="pricing" class="border-y border-[#0b1424]/8 bg-[#f7f8fa]">
      <div class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-28">
        <LandingSectionHeading
          eyebrow="Pricing"
          title="Priced per workspace, not per parcel."
          lede="Start on one counter, add branches when they pay for themselves. Every plan includes the full custody trail — the difference is scale, and how much of the network you run."
          align="center"
        />

        <div class="mt-12">
          <LandingPricingCards />
        </div>
      </div>
    </section>

    <!-- ── FAQ ──────────────────────────────────────────────────────────── -->
    <section id="faq" class="mx-auto max-w-[1200px] px-5 py-20 sm:px-8 sm:py-28">
      <div class="grid gap-10 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] lg:gap-16">
        <LandingSectionHeading
          eyebrow="Questions"
          title="The things operators ask first."
          lede="If your question is not here, the fastest answer is usually to put one branch on it and see."
        />

        <LandingFaqAccordion />
      </div>
    </section>

    <!-- ── Final CTA ────────────────────────────────────────────────────── -->
    <section class="mx-auto max-w-[1200px] px-5 pb-20 sm:px-8 sm:pb-28">
      <div class="relative overflow-hidden rounded-[20px] bg-[#0b1424] px-6 py-14 text-center sm:px-12 sm:py-20">
        <div
          class="pointer-events-none absolute inset-0 bg-[radial-gradient(600px_320px_at_50%_0%,rgba(38,116,218,0.28),transparent_70%)]"
          aria-hidden="true"
        />
        <div
          class="pointer-events-none absolute inset-0 opacity-[0.14] [background-image:linear-gradient(to_right,rgba(255,255,255,0.35)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.35)_1px,transparent_1px)] [background-size:56px_56px] [mask-image:radial-gradient(500px_300px_at_50%_0%,black,transparent)]"
          aria-hidden="true"
        />

        <div class="relative mx-auto max-w-2xl">
          <h2
            class="text-[clamp(1.75rem,4vw,2.9rem)] leading-[1.08] font-semibold tracking-[-0.035em] text-white text-balance"
          >
            Put every parcel on the record.
          </h2>
          <p class="mx-auto mt-4 max-w-xl text-[15.5px] leading-[1.65] text-white/55 text-pretty">
            Start with one branch and the parcels you already move. Add the rest of the network when you are
            ready — the record follows automatically.
          </p>
          <div class="mt-8 flex flex-col items-center justify-center gap-2.5 sm:flex-row">
            <LandingActionLink to="/sign-in" variant="light" size="lg" class="w-full sm:w-auto">
              Get started free
              <UiIcon name="arrowRight" :size="16" />
            </LandingActionLink>
            <LandingActionLink to="#platform" variant="quiet" size="lg" class="w-full text-white/70 hover:text-white sm:w-auto">
              Explore the platform
            </LandingActionLink>
          </div>
          <p class="mt-5 text-[12px] text-white/35">No card required. Sign in with the details your admin gave you.</p>
        </div>
      </div>
    </section>
  </div>
</template>
