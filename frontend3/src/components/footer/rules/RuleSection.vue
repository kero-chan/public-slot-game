<template>
  <div class="rule-section">
    <h4>{{ title }}</h4>
    <div class="section-content">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
}

defineProps<Props>()
</script>

<style scoped lang="scss">
.rule-section {
  margin-bottom: 2rem;
  
  h4 {
    color: #ffd700;
    font-size: 1.2rem;
    margin-bottom: 1rem;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
    border-bottom: 1px solid rgba(255, 215, 0, 0.2);
    padding-bottom: 0.5rem;
  }
}

.section-content {
  color: #e0e0e0;
  line-height: 1.6;
  
  :deep(ul) {
    list-style: none;
    padding: 0;
    
    li {
      padding: 0.5rem 0;
      position: relative;
      padding-left: 1.5rem;
      
      &:before {
        content: "→";
        position: absolute;
        left: 0;
        color: #ffd700;
        font-weight: bold;
      }
    }
  }
  
  :deep(p) {
    margin-bottom: 1rem;
  }
}
</style>