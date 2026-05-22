/**
 * goview.js — the entire JS runtime for goview applications.
 *
 * Responsibilities:
 *   1. __goview.morph()   — called by Go to patch the DOM (via Idiomorph)
 *   2. w-call delegation  — routes user interactions to Go bindings
 *
 * Dependencies: Idiomorph (vendored as idiomorph.min.js, loaded before this file)
 *
 * ~150 lines. No framework. No build step. No npm.
 */

(function () {
  "use strict";

  // ─── DOM primitives called by Go via WindowExecJS / Eval ──────────────────

  window.__goview = {
    /**
     * Morph the innerHTML of the element matching `selector` to `html`.
     * Uses Idiomorph to preserve input focus, values, and scroll position.
     */
    morph(selector, html) {
      const el = document.querySelector(selector);
      if (!el) {
        console.warn(`[goview] selector not found: "${selector}"`);
        return;
      }
      Idiomorph.morph(el, html, { morphStyle: "innerHTML" });
    },

    show(selector) {
      const el = document.querySelector(selector);
      if (el) el.style.display = "";
    },

    hide(selector) {
      const el = document.querySelector(selector);
      if (el) el.style.display = "none";
    },

    addClass(selector, cls) {
      document.querySelector(selector)?.classList.add(cls);
    },

    removeClass(selector, cls) {
      document.querySelector(selector)?.classList.remove(cls);
    },
  };

  // ─── Argument parsing ──────────────────────────────────────────────────────

  function parseArgs(el) {
    const json = el.getAttribute("w-args-json");
    const raw  = el.getAttribute("w-args");
    if (json) {
      try { return [JSON.parse(json)]; } catch (e) {
        console.error("[goview] invalid w-args-json:", json, e);
        return [];
      }
    }
    if (raw !== null) return [raw];
    return [];
  }

  // ─── Go binding call ───────────────────────────────────────────────────────

  function callBinding(el) {
    const method = el.getAttribute("w-call");
    if (!method) return;

    const fn = window[method];
    if (typeof fn !== "function") {
      console.error(`[goview] no binding found for "${method}". Did you call w.Bind("${method}", ...)?`);
      return;
    }

    const args = parseArgs(el);

    // w-value: dynamically read an element's value as the argument
    const valueAttr = el.getAttribute("w-value");
    if (valueAttr !== null) {
      const src = valueAttr === "" ? el : document.querySelector(valueAttr);
      if (src) {
        if (args.length === 0) args.push("");
        args[0] = src.value;
      }
    }

    fn(...args);

    // w-clear: reset an element's value after the call
    const clearAttr = el.getAttribute("w-clear");
    if (clearAttr !== null) {
      const target = clearAttr === "" ? el : document.querySelector(clearAttr);
      if (target) {
        target.value = "";
        if (target.tagName === "INPUT" || target.tagName === "TEXTAREA") {
          target.focus();
        }
      }
    }
  }

  // ─── Click delegation ──────────────────────────────────────────────────────
  // Works on dynamically rendered content because it listens at the document level.

  document.addEventListener("click", function (e) {
    const el = e.target.closest("[w-call]");
    if (!el) return;

    // skip elements with a non-click w-trigger
    const trigger = el.getAttribute("w-trigger");
    if (trigger && trigger !== "click") return;

    e.preventDefault();
    callBinding(el);
  });

  // ─── Input / change / keydown delegation ──────────────────────────────────

  const debounceTimers = new WeakMap();

  function handleTrigger(e) {
    const el = e.target.closest("[w-call][w-trigger]");
    if (!el) return;

    const trigger = el.getAttribute("w-trigger");
    if (trigger !== e.type) return;

    // w-key: only fire on matching keyboard keys
    const keyFilter = el.getAttribute("w-key");
    if (keyFilter && e.key !== keyFilter) return;

    const delay = parseInt(el.getAttribute("w-debounce") || "0", 10);

    if (delay > 0) {
      clearTimeout(debounceTimers.get(el));
      debounceTimers.set(
        el,
        setTimeout(() => callBinding(el), delay)
      );
    } else {
      callBinding(el);
    }
  }

  document.addEventListener("input",  handleTrigger);
  document.addEventListener("change", handleTrigger);
  document.addEventListener("keydown", handleTrigger);

  // ─── Ready log ────────────────────────────────────────────────────────────

  console.log("[goview] runtime ready");

  // ─── DOM ready notification ───────────────────────────────────────────────
  // Notifies Go that the runtime is loaded and the DOM is ready, so Go can
  // mount components. This replaces the manual polling shim in user code.
  //
  // Some webview backends (e.g. webui) inject Go bindings asynchronously,
  // so we retry with a bounded interval instead of polling forever.

  function notifyReady() {
    if (typeof window.__goviewReady === "function") {
      console.log("[goview] DOM ready, calling __goviewReady");
      window.__goviewReady();
      return;
    }

    let attempts = 0;
    const maxAttempts = 100; // 5 seconds at 50 ms intervals
    const interval = setInterval(function () {
      if (typeof window.__goviewReady === "function") {
        clearInterval(interval);
        console.log("[goview] DOM ready, calling __goviewReady (after " + attempts + " retries)");
        window.__goviewReady();
      } else if (++attempts >= maxAttempts) {
        clearInterval(interval);
        console.error(
          "[goview] __goviewReady never became available after " + maxAttempts + " attempts. " +
          "Did you call app.Run()?"
        );
      }
    }, 50);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", notifyReady);
  } else {
    notifyReady();
  }
})();
