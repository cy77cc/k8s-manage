import { useEffect } from 'react';

interface UseAIChatShortcutsOptions {
  draft: string;
  disabled?: boolean;
  onCreateConversation?: () => void;
  onDraftChange: (value: string) => void;
  onSend: (value: string) => void;
}

function isEditableTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false;
  }
  if (target.isContentEditable) {
    return true;
  }
  const tag = target.tagName.toLowerCase();
  return tag === 'input' || tag === 'textarea' || tag === 'select';
}

function focusComposer(): void {
  const composer = document.querySelector<HTMLElement>('[data-ai-chat-composer="true"] textarea');
  if (composer) {
    composer.focus();
    return;
  }
  const fallback = document.querySelector<HTMLElement>('[data-ai-chat-composer="true"] [contenteditable="true"]');
  fallback?.focus();
}

export function useAIChatShortcuts({
  draft,
  disabled,
  onCreateConversation,
  onDraftChange,
  onSend,
}: UseAIChatShortcutsOptions) {
  useEffect(() => {
    if (disabled) {
      return undefined;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.defaultPrevented) {
        return;
      }

      const editable = isEditableTarget(event.target);
      const key = event.key.toLowerCase();

      if ((event.metaKey || event.ctrlKey) && key === 'enter') {
        const trimmed = draft.trim();
        if (!trimmed) {
          return;
        }
        event.preventDefault();
        onSend(trimmed);
        return;
      }

      if (!editable && key === '/') {
        event.preventDefault();
        focusComposer();
        return;
      }

      if (!editable && !event.metaKey && !event.ctrlKey && !event.altKey && key === 'n') {
        event.preventDefault();
        onCreateConversation?.();
        return;
      }

      if (key === 'escape') {
        if (editable) {
          event.preventDefault();
          onDraftChange('');
          if (event.target instanceof HTMLElement) {
            event.target.blur();
          }
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [disabled, draft, onCreateConversation, onDraftChange, onSend]);
}
