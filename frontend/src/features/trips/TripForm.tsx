import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

/** Zod schema for the new-trip form. Drives both validation and TypeScript types. */
const tripFormSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .transform((s) => s.trim()),
  start_date: z
    .string()
    .min(1, 'Start date is required')
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format'),
  // RHF registers empty inputs as '' not undefined; transform to match API expectation.
  end_date: z
    .string()
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
    .transform((v) => v === '' ? undefined : v)
    .optional()
    .or(z.literal('').transform(() => undefined)),
})

/** The validated and coerced form values — inferred directly from the schema. */
export type TripFormValues = z.infer<typeof tripFormSchema>

/** Props for {@link TripForm}. */
interface TripFormProps {
  /** Called with validated form values when the user submits. */
  onSubmit: (values: TripFormValues) => void
  /**
   * When true the submit button is disabled and shows a "Saving…" label.
   * Controlled by the parent so the form stays unaware of async state.
   */
  isSubmitting: boolean
}

/**
 * Presentational form for creating a new trip.
 *
 * Validation is handled entirely by the Zod schema via React Hook Form's
 * zodResolver — the same Zod library used for API response validation, so
 * there's only one validation paradigm in the project.
 *
 * This component owns no server state. Mutations live in the parent
 * (TripsPage) where TanStack Query provides loading/error state.
 */
export function TripForm({ onSubmit, isSubmitting }: TripFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TripFormValues>({
    resolver: zodResolver(tripFormSchema),
  })

  function handleValidSubmit(values: TripFormValues) {
    onSubmit(values)
    reset()
  }

  return (
    <form
      onSubmit={handleSubmit(handleValidSubmit)}
      noValidate
      className="mb-8 space-y-4"
    >
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-gray-700">
          Trip Name
        </label>
        <input
          id="name"
          type="text"
          {...register('name')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
          placeholder="e.g. Pacific Coast 2024"
        />
        {errors.name && (
          <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="start_date" className="block text-sm font-medium text-gray-700">
          Start Date
        </label>
        <input
          id="start_date"
          type="text"
          placeholder="YYYY-MM-DD"
          {...register('start_date')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
        />
        {errors.start_date && (
          <p className="mt-1 text-sm text-red-600">{errors.start_date.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="end_date" className="block text-sm font-medium text-gray-700">
          End Date <span className="text-gray-400">(optional)</span>
        </label>
        <input
          id="end_date"
          type="text"
          placeholder="YYYY-MM-DD"
          {...register('end_date')}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? 'Saving…' : 'Add Trip'}
      </button>
    </form>
  )
}
