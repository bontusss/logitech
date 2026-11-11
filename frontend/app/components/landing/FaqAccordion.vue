<script setup lang="ts">
const items = [
  {
    q: 'Is this a tracking page or an operating system?',
    a: 'An operating system. Tracking is one output of it: every waybill, bag, container and hand-over is recorded at the counter, on the road and at the receiving branch, so the status your customer sees comes from the same record your staff work on.',
  },
  {
    q: 'What happens when only some parcels in a bag are ready?',
    a: 'You dispatch the bag without them. Held parcels keep their own status and reason, and the manifest records exactly which parcels left, in which bag, on which vehicle and with which staff member.',
  },
  {
    q: 'Can we run more than one branch on the same account?',
    a: 'Yes. Branches share one record set, so a parcel sent at one counter is visible to the branch receiving it. Owners and managers see every branch; staff see the counter they work at.',
  },
  {
    q: 'Do we need internet at every counter?',
    a: 'You need a connection to record movements. The app installs like a normal web app on phones and tablets, so counters can keep it open on the device they already use.',
  },
  {
    q: 'Can we put our own logo and domain on it?',
    a: 'Yes, on the Business plan. Each workspace gets its own branding, its own branches, its own staff roles and its own price book, while the data stays isolated from every other tenant.',
  },
  {
    q: 'How does pricing work as we grow?',
    a: 'Plans are configured per workspace, so branches, staff seats and modules are agreed before you commit. The numbers on this page are the starting point, not a lock-in.',
  },
]

const open = ref<number | null>(0)

function toggle(index: number) {
  open.value = open.value === index ? null : index
}
</script>

<template>
  <div class="divide-y divide-[#0b1424]/10 border-y border-[#0b1424]/10">
    <div v-for="(item, index) in items" :key="item.q">
      <h3>
        <button
          :id="`faq-button-${index}`"
          type="button"
          class="flex w-full items-start justify-between gap-4 py-5 text-left transition hover:text-[#2674da] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#2674da]"
          :aria-expanded="open === index"
          :aria-controls="`faq-panel-${index}`"
          @click="toggle(index)"
        >
          <span class="text-[15px] font-medium tracking-[-0.01em] text-[#0b1424]">{{ item.q }}</span>
          <span
            class="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full border border-[#0b1424]/10 text-[#0b1424]/50 transition-transform duration-200"
            :class="open === index ? 'rotate-180' : ''"
            aria-hidden="true"
          >
            <UiIcon name="chevronDown" :size="13" />
          </span>
        </button>
      </h3>
      <div
        v-show="open === index"
        :id="`faq-panel-${index}`"
        role="region"
        :aria-labelledby="`faq-button-${index}`"
        class="pr-10 pb-6"
      >
        <p class="max-w-[62ch] text-[14px] leading-[1.7] text-[#0b1424]/60">{{ item.a }}</p>
      </div>
    </div>
  </div>
</template>
