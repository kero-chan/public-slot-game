<template>
  <div class="history-content">
    <!-- Header Section -->
    <div class="history-header">
      <div class="header-title">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
        </svg>
        <span>Spin History</span>
      </div>
      <div class="total-spins" v-if="totalRecords > 0">
        {{ totalRecords }} spins
      </div>
    </div>
    
    <!-- Loading State -->
    <div v-if="loading && spinHistory.length === 0" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading your spin history...</p>
    </div>
    
    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <div class="error-icon">⚠️</div>
      <p>{{ error }}</p>
      <button @click="fetchSpinHistory()" class="retry-btn">Try Again</button>
    </div>
    
    <!-- Empty State -->
    <div v-else-if="!loading && spinHistory.length === 0" class="empty-state">
      <div class="empty-icon">🎰</div>
      <h3>No Spin History</h3>
      <p>Start playing to see your betting history here!</p>
    </div>
    
    <!-- History List -->
    <div v-else class="history-list" ref="historyListRef">
      <div 
        v-for="(record, index) in spinHistory" 
        :key="record.spin_id" 
        class="history-card"
      >
        <div class="card-header">
          <div class="spin-number">
            <span class="spin-label">Spin</span>
            <span class="spin-value">#{{ index + 1 }}</span>
          </div>
          <div class="spin-time">{{ formatTime(record.created_at) }}</div>
        </div>
        
        <div class="card-content">
          <div class="bet-section">
            <div class="amount-item">
              <div class="amount-label">Bet</div>
              <div class="amount-value bet-amount">${{ formatCurrency(record.bet_amount) }}</div>
            </div>
            
            <div class="amount-item">
              <div class="amount-label">Win</div>
              <div class="amount-value win-amount" :class="{ 
                'positive': record.total_win > 0, 
                'zero': record.total_win === 0 
              }">
                ${{ formatCurrency(record.total_win) }}
              </div>
            </div>
          </div>
          
          <div class="profit-section">
            <div class="profit-label">Net Result</div>
            <div class="profit-value" :class="getNetResultClass(record)">
              {{ formatNetResult(record) }}
            </div>
          </div>
        </div>
        
        <!-- Special indicators -->
        <div class="card-footer" v-if="record.is_free_spin || record.free_spins_triggered">
          <div class="badge free-spin" v-if="record.is_free_spin">Free Spin</div>
          <div class="badge bonus-trigger" v-if="record.free_spins_triggered">Bonus Triggered</div>
        </div>
      </div>
      
      <!-- Load More Section -->
      <div 
        v-if="hasMore && !loadingMore" 
        ref="loadMoreTrigger" 
        class="load-more-section"
      >
        <div class="load-more-content">
          <div class="dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
          <p>Scroll for more</p>
        </div>
      </div>
      
      <!-- Loading More -->
      <div v-if="loadingMore" class="loading-more-section">
        <div class="loading-spinner small"></div>
        <span>Loading more spins...</span>
      </div>
      
      <!-- End of List -->
      <div v-if="!hasMore && spinHistory.length > 0" class="end-section">
        <div class="end-icon">🏁</div>
        <p>You've reached the end</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useIntersectionObserver } from '@vueuse/core'
import { gameApi } from '@/api/game'
import type { SpinSummary, SpinHistoryResponse } from '@/types/api'

// State variables
const spinHistory = ref<SpinSummary[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const error = ref<string | null>(null)
const currentPage = ref(1)
const pageLimit = ref(20)
const totalRecords = ref(0)
const hasMore = ref(true)

// Template refs
const historyListRef = ref<HTMLElement>()
const loadMoreTrigger = ref<HTMLElement>()

// Methods
const fetchSpinHistory = async (page: number = 1, append: boolean = false) => {
  try {
    if (append) {
      loadingMore.value = true
    } else {
      loading.value = true
      spinHistory.value = []
    }
    error.value = null
    
    const response: SpinHistoryResponse = await gameApi.getSpinHistory(page, pageLimit.value)
    
    if (append) {
      // Append new records to existing list
      spinHistory.value = [...spinHistory.value, ...response.spins]
    } else {
      // Replace the list with new records
      spinHistory.value = response.spins
    }
    
    currentPage.value = response.page
    totalRecords.value = response.total
    
    // Check if there are more records to load
    const loadedRecords = spinHistory.value.length
    hasMore.value = loadedRecords < totalRecords.value
    
  } catch (err: any) {
    console.error('Failed to fetch spin history:', err)
    error.value = err.message || 'Failed to load spin history'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

const loadMore = async () => {
  if (loadingMore.value || !hasMore.value) return
  
  const nextPage = currentPage.value + 1
  await fetchSpinHistory(nextPage, true)
}

const formatCurrency = (amount: number): string => {
  return amount.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const formatTime = (timestamp: string) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString([], { 
    day: '2-digit', 
    month: '2-digit', 
    year: '2-digit',
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const formatNetResult = (record: SpinSummary): string => {
  let netValue: number
  
  if (record.is_free_spin) {
    // For free spins, net result is just the win amount
    netValue = record.total_win
  } else {
    // For regular spins, net result is win - bet
    netValue = record.total_win - record.bet_amount
  }
  
  const prefix = netValue > 0 ? '+' : ''
  return `${prefix}$${formatCurrency(netValue)}`
}

const getNetResultClass = (record: SpinSummary): Record<string, boolean> => {
  let netValue: number
  
  if (record.is_free_spin) {
    netValue = record.total_win
  } else {
    netValue = record.total_win - record.bet_amount
  }
  
  return {
    'profit': netValue > 0,
    'loss': netValue < 0,
    'break-even': netValue === 0
  }
}

// Setup intersection observer for infinite scroll
useIntersectionObserver(
  loadMoreTrigger,
  ([{ isIntersecting }]) => {
    if (isIntersecting && hasMore.value && !loadingMore.value) {
      // Add small delay to avoid rapid multiple calls
      setTimeout(() => {
        if (hasMore.value && !loadingMore.value) {
          loadMore()
        }
      }, 100)
    }
  },
  {
    rootMargin: '50px', // Trigger 50px before reaching the element
    threshold: 0.1,
  }
)

// Load initial data
onMounted(() => {
  fetchSpinHistory()
})
</script>

<style scoped lang="scss">
.history-content {
  padding-top: 28px;
  color: #ffffff;
  position: relative;
  min-height: 200px;
  font-size: 14px;
}

// Header Section
.history-header {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.15) 0%, rgba(255, 165, 0, 0.1) 100%);
  border-radius: 8px;
  margin-bottom: 12px;
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 215, 0, 0.2);
  box-shadow: 0 4px 20px rgba(255, 215, 0, 0.1);
  
  .header-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 700;
    color: #ffd700;
    
    svg {
      width: 16px;
      height: 16px;
      opacity: 0.9;
    }
  }
  
  .total-spins {
    background: rgba(255, 215, 0, 0.2);
    color: #ffd700;
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 10px;
    font-weight: 600;
    border: 1px solid rgba(255, 215, 0, 0.3);
  }
}

// Loading States
.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px 12px;
  text-align: center;
  
  p {
    margin-top: 8px;
    color: #b0b0b0;
    font-size: 12px;
  }
}

.loading-state .loading-spinner {
  width: 40px;
  height: 40px;
}

.error-state {
  .error-icon {
    font-size: 48px;
    margin-bottom: 8px;
  }
  
  .retry-btn {
    margin-top: 20px;
    background: linear-gradient(135deg, #ffd700 0%, #ffb700 100%);
    color: #1a1a2e;
    border: none;
    padding: 12px 24px;
    border-radius: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 25px rgba(255, 215, 0, 0.3);
    }
  }
}

.empty-state {
  .empty-icon {
    font-size: 64px;
    margin-bottom: 16px;
    opacity: 0.7;
  }
  
  h3 {
    color: #ffffff;
    margin: 0 0 8px 0;
    font-size: 20px;
    font-weight: 600;
  }
}

// History List
.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  padding-right: 4px;
  
  // Custom scrollbar
  &::-webkit-scrollbar {
    width: 6px;
  }
  
  &::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 10px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: linear-gradient(135deg, rgba(255, 215, 0, 0.6) 0%, rgba(255, 165, 0, 0.4) 100%);
    border-radius: 10px;
    
    &:hover {
      background: linear-gradient(135deg, rgba(255, 215, 0, 0.8) 0%, rgba(255, 165, 0, 0.6) 100%);
    }
  }
}

// History Cards
.history-card {
  background: linear-gradient(135deg, rgba(40, 40, 70, 0.8) 0%, rgba(30, 30, 50, 0.8) 100%);
  border-radius: 8px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, #ffd700, #ff6b6b, #4ade80);
    opacity: 0;
    transition: opacity 0.3s ease;
  }
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
    border-color: rgba(255, 215, 0, 0.3);
    
    &::before {
      opacity: 1;
    }
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  
  .spin-number {
    display: flex;
    flex-direction: column;
    
    .spin-label {
      font-size: 9px;
      color: #b0b0b0;
      text-transform: uppercase;
      letter-spacing: 0.3px;
    }
    
    .spin-value {
      font-size: 14px;
      font-weight: 700;
      color: #ffd700;
      margin-top: 1px;
    }
  }
  
  .spin-time {
    background: rgba(255, 255, 255, 0.1);
    color: #e0e0e0;
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 500;
  }
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bet-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.amount-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  
  .amount-label {
    font-size: 9px;
    color: #b0b0b0;
    text-transform: uppercase;
    letter-spacing: 0.3px;
    margin-bottom: 4px;
  }
  
  .amount-value {
    font-size: 12px;
    font-weight: 700;
    
    &.bet-amount {
      color: #ff9500;
    }
    
    &.win-amount {
      &.positive {
        color: #4ade80;
      }
      
      &.zero {
        color: #9ca3af;
      }
    }
  }
}

.profit-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  
  .profit-label {
    font-size: 10px;
    color: #b0b0b0;
    font-weight: 500;
  }
  
  .profit-value {
    font-size: 12px;
    font-weight: 700;
    
    &.profit {
      color: #4ade80;
    }
    
    &.loss {
      color: #ff6b6b;
    }
    
    &.break-even {
      color: #ffd700;
    }
  }
}

.card-footer {
  display: flex;
  gap: 4px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.badge {
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 9px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  
  &.free-spin {
    background: linear-gradient(135deg, #8b5cf6 0%, #6366f1 100%);
    color: #ffffff;
  }
  
  &.bonus-trigger {
    background: linear-gradient(135deg, #ffd700 0%, #ff8c00 100%);
    color: #1a1a2e;
  }
}

// Load More Section
.load-more-section {
  display: flex;
  justify-content: center;
  padding: 16px;
  
  .load-more-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    
    .dots {
      display: flex;
      gap: 4px;
      
      span {
        width: 6px;
        height: 6px;
        background: rgba(255, 215, 0, 0.6);
        border-radius: 50%;
        animation: pulse 1.5s infinite ease-in-out;
        
        &:nth-child(1) { animation-delay: 0s; }
        &:nth-child(2) { animation-delay: 0.3s; }
        &:nth-child(3) { animation-delay: 0.6s; }
      }
    }
    
    p {
      color: #b0b0b0;
      font-size: 11px;
      margin: 0;
    }
  }
}

.loading-more-section {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 12px;
  color: #ffd700;
  
  span {
    font-size: 11px;
    font-weight: 500;
  }
}

.end-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  margin-top: 12px;
  
  .end-icon {
    font-size: 24px;
    margin-bottom: 6px;
    opacity: 0.7;
  }
  
  p {
    color: #b0b0b0;
    font-size: 11px;
    margin: 0;
  }
}

// Loading Spinner
.loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 215, 0, 0.2);
  border-top: 2px solid #ffd700;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  
  &.small {
    width: 16px;
    height: 16px;
    border-width: 2px;
  }
}

// Animations
@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.4;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

// Mobile Responsive
@media (max-width: 768px) {
  .history-content {
    padding: 8px;
    border-radius: 8px;
    font-size: 12px;
  }
  
  .history-header {
    padding: 6px 8px;
    margin-bottom: 8px;
    
    .header-title {
      font-size: 12px;
      gap: 6px;
      
      svg {
        width: 14px;
        height: 14px;
      }
    }
    
    .total-spins {
      font-size: 9px;
      padding: 3px 6px;
    }
  }
  
  .history-card {
    padding: 8px;
  }
  
  .card-header {
    .spin-value {
      font-size: 12px;
    }
    
    .spin-time {
      font-size: 9px;
    }
  }
  
  .amount-item {
    padding: 6px;
    
    .amount-value {
      font-size: 11px;
    }
  }
  
  .profit-section {
    padding: 6px;
    
    .profit-value {
      font-size: 11px;
    }
  }
}

@media (max-width: 480px) {
  .bet-section {
    grid-template-columns: 1fr;
    gap: 6px;
  }
  
  .profit-section {
    flex-direction: column;
    gap: 4px;
    text-align: center;
  }
  
  .history-card {
    padding: 6px;
  }
  
  .amount-value,
  .profit-value {
    font-size: 10px !important;
  }
}
</style>