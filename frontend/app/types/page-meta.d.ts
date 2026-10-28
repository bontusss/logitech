declare module '#app' {
  interface PageMeta {
    /** When set, only these roles may open the page. */
    roles?: string[]
  }
}

export {}
