// Extends Vitest's expect() with DOM matchers like toBeInTheDocument(),
// toHaveTextContent(), etc. — provided by @testing-library/jest-dom.
// This file is loaded before every test via vite.config.ts setupFiles.
import '@testing-library/jest-dom'
