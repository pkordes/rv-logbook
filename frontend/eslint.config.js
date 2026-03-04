import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import { defineConfig, globalIgnores } from 'eslint/config'

/**
 * Custom ESLint plugin: every <button> must have a stable E2E test identifier.
 *
 * A stable identifier is one of:
 *   - aria-label  → preferred when the action has inherent meaning
 *                   (e.g. aria-label="Delete camping" on a per-row delete button)
 *   - data-testid → for structural/state-changing elements where aria-label
 *                   doesn't apply naturally
 *                   (e.g. data-testid="trip-form-submit" on a submit button
 *                   whose text flips "Add Trip" → "Saving…")
 *
 * CSS class names must NEVER be used as Playwright selectors — they belong to
 * styling and must not be coupled to test contracts.
 *
 * See CONTRIBUTING.md § "E2E Testability" for the full selector strategy.
 */
const testabilityPlugin = {
  rules: {
    'interactive-has-test-id': {
      meta: {
        type: 'problem',
        docs: {
          description:
            'Require aria-label or data-testid on all <button> elements for stable E2E targeting.',
        },
        messages: {
          missing:
            '<button> has no stable E2E identifier. ' +
            'Add aria-label="{description}" for semantically meaningful actions ' +
            '(e.g. aria-label="Delete camping"), or data-testid="{resource}-{element}" ' +
            'for structural/state-changing buttons (e.g. data-testid="trip-form-submit"). ' +
            'See CONTRIBUTING.md § "E2E Testability".',
        },
      },
      create(context) {
        return {
          JSXOpeningElement(node) {
            if (
              node.name.type !== 'JSXIdentifier' ||
              node.name.name !== 'button'
            )
              return

            const attrs = node.attributes

            const hasAriaLabel = attrs.some(
              (attr) =>
                attr.type === 'JSXAttribute' &&
                attr.name.type === 'JSXIdentifier' &&
                attr.name.name === 'aria-label' &&
                attr.value !== null,
            )
            const hasTestId = attrs.some(
              (attr) =>
                attr.type === 'JSXAttribute' &&
                attr.name.type === 'JSXIdentifier' &&
                attr.name.name === 'data-testid' &&
                attr.value !== null,
            )

            if (!hasAriaLabel && !hasTestId) {
              context.report({ node, messageId: 'missing' })
            }
          },
        }
      },
    },
  },
}

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
      jsxA11y.flatConfigs.recommended,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    rules: {
      // autoFocus on inline-edit inputs is a deliberate accessibility improvement
      // (keyboard users benefit from focus moving to the newly active field).
      // The jsx-a11y recommendation to avoid it is a general heuristic that does
      // not apply to this specific, intentional UX pattern.
      'jsx-a11y/no-autofocus': 'off',
    },
  },
  {
    // Custom testability rule: applies only to component source files.
    // Excluded from test files (*.test.tsx) where we query elements without testids.
    files: ['src/**/*.{ts,tsx}'],
    ignores: ['src/**/*.test.{ts,tsx}', 'src/test/**'],
    plugins: { testability: testabilityPlugin },
    rules: {
      'testability/interactive-has-test-id': 'error',
    },
  },
])
