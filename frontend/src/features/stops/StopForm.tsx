import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import type { Stop } from '../../api/stops'
import { TagInput } from '../../components/TagInput'

/**
 * Converts a YYYY-MM-DD date string to an RFC 3339 timestamp representing
 * noon Eastern Standard Time (UTC−5 = T17:00:00Z).
 *
 * Why noon EST?
 * JavaScript's `new Date("YYYY-MM-DD")` treats plain date strings as UTC midnight
 * (T00:00:00Z). In any US timezone that moment falls on the *previous* calendar
 * day (e.g. midnight UTC = 7 PM EST the night before). Using noon EST (17:00 UTC)
 * means the stored instant is solidly within the entered day from Hawaii (UTC−10,
 * 7 AM) to Maine (UTC−5, 12 PM) and will never roll back to the day before when
 * displayed in any US locale.
 */
const dateToRfc3339 = (val: string) => `${val}T17:00:00Z`

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
  /** Pending tag names collected via TagInput. */
  tags: z.array(z.string()).default([]),
})

type StopFormInput = z.input<typeof stopFormSchema>
type StopFormOutput = z.output<typeof stopFormSchema>

/**
 * Validated stop form values passed to the onSubmit callback.
 * `tagNames` is the list of tag name strings the user added via TagInput.
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
  /**
   * When provided the form is pre-filled and the submit button reads
   * "Save Changes" instead of "Add Stop".  The arrived_at timestamp is
   * converted back to YYYY-MM-DD for display in the text input.
   */
  initialValues?: Stop
  /** Called when the user clicks Cancel (only rendered when initialValues is set). */
  onCancel?: () => void
}

/**
 * Presentational form for adding or editing a stop on a trip.
 *
 * Uses {@link TagInput} for tag entry — pills are shown as the user types,
 * with autocomplete from existing tags. The validated `tagNames` array is
 * passed to the parent via `onSubmit`.
 */
export function StopForm({ onSubmit, isSubmitting, initialValues, onCancel }: StopFormProps) {
  const isEditing = Boolean(initialValues)
  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<StopFormInput>({
    resolver: zodResolver(stopFormSchema),
    defaultValues: initialValues
      ? {
          name: initialValues.name,
          // arrived_at is stored as RFC 3339 — slice to YYYY-MM-DD for the date text input
          arrived_at: initialValues.arrived_at.slice(0, 10),
          departed_at: initialValues.departed_at?.slice(0, 10) ?? '',
          location: initialValues.location ?? '',
          notes: initialValues.notes ?? '',
          tags: initialValues.tags.map((t) => t.name),
        }
      : undefined,
  })

  // When the edit target changes (user switches from one stop to another),
  // reset the form to the new values. Also handles the case where the form
  // is rendered in a parent that mounts it before RHF's ref callbacks fire.
  useEffect(() => {
    if (initialValues) {
      reset({
        name: initialValues.name,
        arrived_at: initialValues.arrived_at.slice(0, 10),
        departed_at: initialValues.departed_at?.slice(0, 10) ?? '',
        location: initialValues.location ?? '',
        notes: initialValues.notes ?? '',
        tags: initialValues.tags.map((t) => t.name),
      })
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialValues?.id])

  function handleValidSubmit(values: StopFormOutput) {
    onSubmit({ ...values, tagNames: values.tags })
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
        {/* Using <span> instead of <label> because TagInput manages its own
            aria-label ("Add tag") internally — a floating <label> with no
            htmlFor association would be an a11y violation. */}
        <span className="block text-sm font-medium text-gray-700">
          Tags
        </span>
        <Controller
          name="tags"
          control={control}
          render={({ field }) => (
            <TagInput value={field.value ?? []} onChange={field.onChange} />
          )}
        />
      </div>

      <button
        type="submit"
        data-testid="stop-form-submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? 'Saving…' : isEditing ? 'Save Changes' : 'Add Stop'}
      </button>
      {isEditing && onCancel && (
        <button
          type="button"
          aria-label="Cancel editing stop"
          onClick={onCancel}
          className="ml-2 rounded border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
        >
          Cancel
        </button>
      )}
    </form>
  )
}
