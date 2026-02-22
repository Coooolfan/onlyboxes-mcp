import { onBeforeUnmount, onMounted, type Ref } from 'vue'

type UseDismissibleMenuOptions = {
  containerRef: Ref<HTMLElement | null>
  isOpen: Ref<boolean>
  onClose: () => void
}

export function useDismissibleMenu({
  containerRef,
  isOpen,
  onClose,
}: UseDismissibleMenuOptions): void {
  function handlePointerDown(event: Event): void {
    if (!isOpen.value) {
      return
    }

    const target = event.target
    if (!(target instanceof Node)) {
      return
    }
    if (containerRef.value?.contains(target)) {
      return
    }
    onClose()
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Escape' || !isOpen.value) {
      return
    }
    onClose()
  }

  onMounted(() => {
    if (typeof document === 'undefined') {
      return
    }
    document.addEventListener('pointerdown', handlePointerDown)
    document.addEventListener('keydown', handleKeydown)
  })

  onBeforeUnmount(() => {
    if (typeof document === 'undefined') {
      return
    }
    document.removeEventListener('pointerdown', handlePointerDown)
    document.removeEventListener('keydown', handleKeydown)
  })
}
