import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

/**
 * Converts a YYYY-MM-DD date string to a midnight UTC RFC 3339 timestamp.
 * The backend requires date-time format; for a logbook the day is what matters.
 */
const dateToRfc3339 = (val: string) => `${val}T00:00:00Z`

/** Internal Zod schema — validates raw form field strings. */
const stopFormSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .transform((s) => s.trim()),
  arrived_at: z
    .string()
    .min(1, 'Arrived at is required')
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
    .transform(dateToRfc3339),
  departed_at: z
    .string()
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
    .transform(dateToRfc3339)
    .optional()
    .or(z.literal('').transform(() => undefined)),
  location: z
    .string()
    .or(z.literal('').transform(() => undefined))
    .optional(),
  notes: z
    .string()
    .or(z.literal('').transform(() => undefined))
    .optional(),
  /** Raw comma-separated tag names entered by the user. */
  tagsRaw: z.string().optional(),
})

type StopFormInput = z.input<typeof stopFormSchema>
type StopFormOutput = z.output<typeof stopFormSchema>

/**
 * Validated stop form values, including the parsed tag name array.
 * The `tagsRaw` field is also present (from the schema output) but consumers
 * should use `tagNames` instead.
 */
export type StopFormValues = StopFormOutput & { tagNames: string[] }

/** Props for {@link StopForm}. */
interface StopFormProps {
  /** Called with validated values (including parsed tagNames) when the user submits. */
  onSubmit: (values: StopFormValues) => void
  /**
   * When true the submit button is disabled and shows "Saving…".
   * Controlled by the parent so the form stays unaware of async state.
   */
  isSubmitting: boolean
}

/**
 * Presentational form for adding a stop to a trip.
 *
 * The tag input accepts a comma-separated list of names (e.g. "camping, national park").
 * Splitting and trimming happens here so callers receive a clean `string[]`.
 */
export function StopForm({ onSubmit, isSubmitting }: StopFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<StopFormInput>({
    resolver: zodResolver(stopFormSchema),
  })

  function handleValidSubmit(values: StopFormOutput) {
    const tagNames = (values.tagsRaw ?? '')
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)
    onSubmit({ ...values, tagNames })
    reset()
  }

  return (
    <form
      onSubmit={handleSubmit(handleValidSubmit as (data: StopFormInput) => void)}
      noValidate
      className="mb-8 space-y-4"
    >
      {/* Name */}
      <div>
        <label htmlFor="stop-name" className="block text-sm font-medium text-gray-700">
          Stop Name
        </label>
        <input
          id="stop-name"
          type="text"
          {...register('name')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="e.g. Yellowstone Camp"
        />
        {errors.name && (
          <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
        )}
      </div>

      {/* Arrived At */}
      <div>
        <label htmlFor="arrived-at" className="block text-sm font-medium text-gray-700">
          Arrived At
        </label>
        <input
          id="arrived-at"
          type="text"
          {...register('arrived_at')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="e.g. 2025-06-02"
        />
        {errors.arrived_at && (
          <p className="mt-1 text-sm text-red-600">{errors.arrived_at.message}</p>
        )}
      </div>

      {/* Departed At */}
      <div>
        <label htmlFor="departed-at" className="block text-sm font-medium text-gray-700">
          Departed At
        </label>
        <input
          id="departed-at"
          type="text"
          {...register('departed_at')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="e.g. 2025-06-04 (optional)"
        />
      </div>

      {/* Location */}
      <div>
        <label htmlFor="location" className="block text-sm font-medium text-gray-700">
          Location
        </label>
        <input
          id="location"
          type="text"
          {...register('location')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="e.g. Yellowstone, WY (optional)"
        />
      </div>

      {/* Notes */}
      <div>
        <label htmlFor="notes" className="block text-sm font-medium text-gray-700">
          Notes
        </label>
        <textarea
          id="notes"
          {...register('notes')}
          rows={2}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="Optional notes"
        />
      </div>

      {/* Tags */}
      <div>
        <label htmlFor="tags-raw" className="block text-sm font-medium text-gray-700">
          Tags (comma-separated)
        </label>
        <input
          id="tags-raw"
          type="text"
          {...register('tagsRaw')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="e.g. camping, national park"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? 'Saving…' : 'Add Stop'}
      </button>
    </form>
  )
}
