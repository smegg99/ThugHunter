/** Scroll the main layout container back to the top. */
export function scrollToTop() {
  document.querySelector('.main-scroll-container')?.scrollTo({ top: 0, behavior: 'smooth' })
}
