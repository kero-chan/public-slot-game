export function formatNumber(num: number): string {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(num)
}

export function formatWinAmount(num: number): string {
  const formatted = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2
  }).format(num)
  
  // Remove unnecessary trailing zeros after decimal point
  return formatted.replace(/\.0+$/, '').replace(/(\.\d*?)0+$/, '$1')
}
