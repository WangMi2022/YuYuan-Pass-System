<template>
  <section class="na-sidebar-tip" aria-label="资产管理小贴士">
    <div class="na-sidebar-tip__content">
      <strong>资产管理小贴士</strong>
      <p>定期核对资产状态，让盘点与维护更高效。</p>
    </div>
    <svg
      class="na-sidebar-tip__illustration"
      viewBox="0 0 120 68"
      role="img"
      aria-label="桌面绿植插图"
    >
      <path class="plant-stem" d="M60 49V18M60 35C49 33 44 27 42 19M60 41C72 39 78 32 80 24M60 28C68 25 72 20 74 14" />
      <path class="plant-leaf plant-leaf--soft" d="M41 18c10 0 16 5 18 15-10 0-16-5-18-15Z" />
      <path class="plant-leaf" d="M82 23c-11 0-18 6-21 17 11 0 18-6 21-17Z" />
      <path class="plant-leaf" d="M75 12c-8 1-13 6-14 15 8-1 13-6 14-15Z" />
      <path class="plant-leaf plant-leaf--soft" d="M51 27c-8 1-13 6-14 14 8-1 13-6 14-14Z" />
      <path class="plant-pot" d="M48 47h24l-3 14H51l-3-14Z" />
      <path class="plant-base" d="M42 62h36" />
    </svg>
  </section>
</template>

<script setup>
  defineOptions({ name: 'SidebarTipCard' })
</script>

<style scoped lang="scss">
  .na-sidebar-tip {
    position: absolute;
    right: 10px;
    bottom: 56px;
    left: 10px;
    height: 116px;
    overflow: hidden;
    border: 1px solid color-mix(in srgb, var(--na-primary) 12%, var(--na-border));
    border-radius: 8px;
    background: color-mix(in srgb, var(--na-primary-soft) 52%, var(--na-card));
    color: var(--na-foreground);
  }
  .na-sidebar-tip__content { position: relative; z-index: 1; width: 128px; padding: 14px 0 0 14px; }
  .na-sidebar-tip__content strong { display: block; font-size: 13px; font-weight: 650; line-height: 1.5; }
  .na-sidebar-tip__content p { margin: 5px 0 0; color: var(--na-muted-foreground); font-size: 11px; line-height: 1.55; }
  .na-sidebar-tip__illustration {
    position: absolute;
    right: 5px;
    bottom: 8px;
    width: 76px;
    height: 54px;
    color: var(--na-primary);
  }
  .plant-stem { fill: none; stroke: currentColor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 2; }
  .plant-leaf { fill: currentColor; opacity: .82; }
  .plant-leaf--soft { opacity: .46; }
  .plant-pot { fill: color-mix(in srgb, var(--na-primary) 18%, var(--na-card)); stroke: color-mix(in srgb, var(--na-primary) 36%, var(--na-border)); stroke-width: 1.5; }
  .plant-base { fill: none; stroke: var(--na-border-strong); stroke-linecap: round; stroke-width: 3; }
  @media (max-height: 680px) { .na-sidebar-tip { display: none; } }
</style>
