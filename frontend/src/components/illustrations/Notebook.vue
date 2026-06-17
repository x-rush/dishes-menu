<script setup lang="ts">
defineProps<{ size?: number }>()
</script>

<template>
  <svg
    :width="size ?? 140"
    :height="size ?? 140"
    viewBox="0 0 140 140"
    fill="none"
    aria-hidden="true"
  >
    <defs>
      <radialGradient id="bgGlow" cx="50%" cy="55%" r="50%">
        <stop offset="0%" stop-color="#ffe0ec" stop-opacity="0.9" />
        <stop offset="100%" stop-color="#ffe0ec" stop-opacity="0" />
      </radialGradient>
      <linearGradient id="paperGrad" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#fff8f3" />
        <stop offset="100%" stop-color="#fff0f5" />
      </linearGradient>
    </defs>
    <circle cx="70" cy="77" r="62" fill="url(#bgGlow)" />

    <rect
      x="34" y="32" width="62" height="80" rx="6"
      fill="url(#paperGrad)"
      stroke="#ffc7dc" stroke-width="1.5"
    />
    <rect x="32" y="32" width="8" height="80" rx="3" fill="#ec7da6" />
    <circle cx="36" cy="44" r="1.6" fill="#fff" />
    <circle cx="36" cy="60" r="1.6" fill="#fff" />
    <circle cx="36" cy="76" r="1.6" fill="#fff" />
    <circle cx="36" cy="92" r="1.6" fill="#fff" />

    <rect x="62" y="26" width="22" height="10" rx="2" fill="#ffb0ce" opacity="0.85" transform="rotate(-6 73 31)" />
    <line x1="62" y1="31" x2="84" y2="31" stroke="#fff" stroke-width="0.6" stroke-dasharray="1.5 1.5" />

    <line x1="50" y1="50" x2="88" y2="50" stroke="#ebc4be" stroke-width="1.2" stroke-linecap="round" />
    <line x1="50" y1="62" x2="84" y2="62" stroke="#ebc4be" stroke-width="1.2" stroke-linecap="round" />
    <line x1="50" y1="74" x2="86" y2="74" stroke="#ebc4be" stroke-width="1.2" stroke-linecap="round" />
    <line x1="50" y1="86" x2="78" y2="86" stroke="#ebc4be" stroke-width="1.2" stroke-linecap="round" />

    <g transform="translate(82, 96)">
      <path
        d="M0 3.5C0 1.6 1.5 0 3.2 0c1 0 1.9 0.4 2.4 1.1 0.5-0.7 1.4-1.1 2.4-1.1 1.7 0 3.2 1.6 3.2 3.5 0 3-5.6 6-5.6 6S0 6.5 0 3.5z"
        fill="#ec7da6"
      />
    </g>

    <g fill="#f7d560" opacity="0.85">
      <path d="M104 32l1 2.4 2.4 0.4-1.8 1.6 0.5 2.4-2.1-1.2-2.1 1.2 0.5-2.4-1.8-1.6 2.4-0.4z" />
      <path d="M28 90l0.7 1.7 1.7 0.3-1.3 1.1 0.4 1.7-1.5-0.8-1.5 0.8 0.4-1.7-1.3-1.1 1.7-0.3z" />
    </g>
  </svg>
</template>
