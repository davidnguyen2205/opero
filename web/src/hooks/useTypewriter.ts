import { useEffect, useRef, useState } from "react";

type UseTypewriterOptions = {
  /** Delay between typing each character (ms). */
  typingSpeed?: number;
  /** How long to keep a fully-typed phrase on screen before clearing (ms). */
  pauseDuration?: number;
};

export type TypewriterPhase = "typing" | "paused";

/** Pick a random index in [0, length), avoiding `exclude` when possible. */
function randomIndex(length: number, exclude = -1): number {
  if (length <= 1) return 0;
  let next = Math.floor(Math.random() * length);
  if (next === exclude) next = (next + 1) % length;
  return next;
}

/**
 * Shows `phrases` in random order, typing each one character-by-character.
 * When a phrase is fully typed it is held for `pauseDuration`, then cleared
 * instantly (no character-by-character delete) before the next random phrase
 * starts typing (never the same one twice in a row). Returns the text
 * currently shown plus the current `phase` (so callers can, e.g., highlight
 * the leading character while it is actively typing).
 *
 * Ported from the Blazeup Super Admin login (useTypewriter.ts).
 */
export function useTypewriter(
  phrases: string[],
  { typingSpeed = 90, pauseDuration = 1800 }: UseTypewriterOptions = {},
) {
  const [text, setText] = useState("");
  const [phase, setPhase] = useState<TypewriterPhase>("typing");
  const phraseIndexRef = useRef(0);
  const textRef = useRef("");

  useEffect(() => {
    if (phrases.length === 0) return;

    // Always lead with the first phrase on mount; randomize from then on.
    phraseIndexRef.current = 0;
    textRef.current = "";
    setText("");
    setPhase("typing");

    let timeoutId: ReturnType<typeof setTimeout>;

    const tick = () => {
      const current = phrases[phraseIndexRef.current % phrases.length];
      const next = current.slice(0, textRef.current.length + 1);

      textRef.current = next;
      setText(next);

      if (next === current) {
        // Fully typed: hold, then clear all at once and move to next phrase.
        setPhase("paused");
        timeoutId = setTimeout(() => {
          textRef.current = "";
          setText("");
          phraseIndexRef.current = randomIndex(phrases.length, phraseIndexRef.current);
          setPhase("typing");
          timeoutId = setTimeout(tick, typingSpeed);
        }, pauseDuration);
      } else {
        setPhase("typing");
        timeoutId = setTimeout(tick, typingSpeed);
      }
    };

    timeoutId = setTimeout(tick, typingSpeed);

    return () => clearTimeout(timeoutId);
  }, [phrases, typingSpeed, pauseDuration]);

  return { text, phase };
}
