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
    right: 12px;
    bottom: 50px;
    left: 12px;
    height: 88px;
    overflow: hidden;
    border: 1px solid var(--sidebar-line);
    border-radius: 6px;
    background: var(--sidebar-surface);
    color: var(--sidebar-text-strong);
    box-shadow: none;
  }
  .na-sidebar-tip::before { position: absolute; inset: 11px auto 11px 0; width: 2px; border-radius: 0 2px 2px 0; background: var(--na-primary); content: ''; }
  .na-sidebar-tip__content { position: relative; z-index: 1; width: calc(100% - 56px); padding: 11px 0 0 13px; }
  .na-sidebar-tip__content strong { display: block; font-size: 11px; font-weight: 650; line-height: 1.5; }
  .na-sidebar-tip__content p { margin: 3px 0 0; color: var(--sidebar-text); font-size: 9px; line-height: 1.55; }
  .na-sidebar-tip__illustration {
    position: absolute;
    right: 4px;
    bottom: 6px;
    width: 60px;
    height: 44px;
    color: color-mix(in srgb, var(--na-primary) 72%, white);
    opacity: .72;
  }
  .plant-stem { fill: none; stroke: currentColor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 2; }
  .plant-leaf { fill: currentColor; opacity: .82; }
  .plant-leaf--soft { opacity: .46; }
  .plant-pot { fill: color-mix(in srgb, var(--na-primary) 22%, var(--sidebar-surface)); stroke: color-mix(in srgb, var(--na-primary) 42%, var(--sidebar-line)); stroke-width: 1.5; }
  .plant-base { fill: none; stroke: var(--sidebar-line); stroke-linecap: round; stroke-width: 3; }
  @media (max-height: 680px) { .na-sidebar-tip { display: none; } }
</style>
