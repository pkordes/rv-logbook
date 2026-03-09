python scripts/spec-format.py --lang ts frontend/src

apiFetch:
 ✔ prepends /api to the path
 ✔ returns parsed JSON on a 2xx response
 ✔ sets Content-Type: application/json by default
 ✔ throws ApiError with the response status on a 4xx response
 ✔ includes the HTTP status code on the thrown ApiError
 ✔ includes the server error message on the thrown ApiError
 ✔ resolves without throwing on a 204 No Content response

fetchExportBlob:
 ✔ calls GET /export with Accept: application/json by default
 ✔ calls GET /export?format=csv with Accept: text/csv when format is csv
 ✔ throws when the server responds with a non-2xx status

listStops:
 ✔ calls GET /api/trips/:id/stops
 ✔ returns a validated StopListResponse

createStop:
 ✔ calls POST /api/trips/:id/stops
 ✔ returns the created stop

deleteStop:
 ✔ calls DELETE /api/trips/:tripId/stops/:stopId

removeTagFromStop:
 ✔ calls DELETE /api/trips/:tripId/stops/:stopId/tags/:slug

searchTags:
 ✔ calls GET /api/tags?q=... with the provided prefix
 ✔ returns the data array (not the paginated wrapper)
 ✔ returns an empty array when q matches nothing

listAllTags:
 ✔ calls GET /api/tags with page and limit params
 ✔ returns a validated TagListResponse
 ✔ forwards custom page and limit to the URL

patchTag:
 ✔ calls PATCH /api/tags/:slug with the new name
 ✔ returns the updated Tag

deleteTag:
 ✔ calls DELETE /api/tags/:slug
 ✔ resolves without a value on 204

createTag:
 ✔ calls POST /api/tags with the tag name
 ✔ returns the created Tag

ErrorBoundary:
 ✔ renders children when there is no error
 ✔ renders the fallback UI when a child throws
 ✔ displays the error message in the fallback UI

TagInput:
 ✔ renders a text input
 ✔ renders existing value entries as removable pills
 ✔ calls onChange without the tag when a pill remove button is clicked
 ✔ calls onChange with new tag appended when Enter is pressed
 ✔ trims the tag name when adding via Enter
 ✔ does not call onChange when Enter is pressed with blank input
 ✔ removes the last tag when Backspace is pressed on an empty input
 ✔ does not call onChange on Backspace when value is empty
 ✔ calls searchTags when the user types 2 or more characters
 ✔ does not call searchTags when fewer than 2 characters are typed
 ✔ renders suggestions in a dropdown when searchTags resolves
 ✔ adds a suggestion to value when clicked and clears the input
 ✔ hides the dropdown after a suggestion is selected
 ✔ does not add a duplicate tag
 ✔ adds tag when Tab is pressed with non-empty input
 ✔ does not add tag when Tab is pressed with empty input

TagPill:
 ✔ renders the tag name
 ✔ does not render a remove button when onRemove is not provided
 ✔ renders a remove button when onRemove is provided
 ✔ calls onRemove when the remove button is clicked

StopForm:
 ✔ renders name, arrived_at, and a submit button
 ✔ shows a validation error when submitted with an empty name
 ✔ shows a validation error when submitted with an empty arrived_at
 ✔ shows a validation error when arrived_at is not in YYYY-MM-DD format
 ✔ calls onSubmit with trimmed values when form is valid
 ✔ adds tags via TagInput and submits them as tagNames
 ✔ calls onSubmit with an empty tagNames array when no tags are added
 ✔ disables the submit button while isSubmitting is true
 ✔ pre-fills the form with initialValues when provided
 ✔ shows a Save Changes button when initialValues is provided
 ✔ calls onCancel when the cancel button is clicked

StopList:
 ✔ shows an empty-state message when there are no stops
 ✔ renders each stop name
 ✔ renders the location when present
 ✔ renders the arrived_at date for each stop
 ✔ calls onDelete with the stop id when the delete button is clicked
 ✔ calls onEdit with the full stop object when the edit button is clicked
 ✔ renders tags as pills when the stop has tags
 ✔ renders no tag pills when the stop has no tags

stopKeys:
 ✔ scopes list key under the trip

useStops:
 ✔ calls listStops with the tripId
 ✔ exposes the stops in the data field

useCreateStop:
 ✔ calls createStop and invalidates the stop list

useDeleteStop:
 ✔ calls deleteStop with both ids

useUpdateStop:
 ✔ calls updateStop with tripId, stopId, and input

tagKeys:
 ✔ provides a stable list key

useTags:
 ✔ calls listAllTags with default page and limit
 ✔ exposes data from the response

useUpdateTag:
 ✔ calls patchTag and invalidates the tag list

useDeleteTag:
 ✔ calls deleteTag and invalidates the tag list
 ✔ also invalidates stop queries so cached trip pages refresh

useCreateTag:
 ✔ calls createTag and invalidates the tag list

ExportButton:
 ✔ renders an Export CSV button
 ✔ calls fetchExportBlob with csv format when clicked
 ✔ creates an object URL from the downloaded blob
 ✔ revokes the object URL after triggering the download
 ✔ disables the button while the download is in flight
 ✔ shows an error message when the download fails

TripForm:
 ✔ renders name, start date, and a submit button
 ✔ shows a validation error when submitted with an empty name
 ✔ shows a validation error when submitted with an empty start date
 ✔ calls onSubmit with trimmed name and start_date when form is valid
 ✔ shows a validation error when start date is not in YYYY-MM-DD format
 ✔ disables the submit button while isSubmitting is true

TripList:
 ✔ shows an empty-state message when there are no trips
 ✔ renders each trip name
 ✔ renders the start date for each trip
 ✔ calls onDelete with the trip id when the delete button is clicked

TripTimeline:
 ✔ renders a timeline entry for each stop
 ✔ renders stops in chronological order regardless of input order
 ✔ renders an empty state when there are no stops
 ✔ renders stop location when provided
 ✔ renders stop tags as pills
 ✔ renders the arrived_at date in a human-readable form
 ✔ handles stops with null arrived_at by placing them at the end

TagsPage:
 ✔ renders the page heading
 ✔ renders a row for each tag showing name and slug
 ✔ shows a loading spinner while data is loading
 ✔ shows an error message on fetch failure
 ✔ shows inline confirmation when Delete is clicked
 ✔ calls deleteTag mutation when Confirm delete is clicked
 ✔ hides confirmation and does not delete when Keep is clicked
 ✔ shows an inline rename form when Edit is clicked
 ✔ calls updateTag mutation with the new name on Save
 ✔ hides the rename form when Cancel is clicked
 ✔ renders the new tag name input
 ✔ calls createTag mutation when the new-tag form is submitted
 ✔ clears the new-tag input after submission

TripDetailPage:
 ✔ renders the trip name as a heading
 ✔ renders the trip start date
 ✔ renders the empty-state stop message when there are no stops
 ✔ renders the add stop form
 ✔ shows a loading spinner while the trip is loading
 ✔ shows an error message when the trip fails to load
 ✔ switches to the edit form when the edit button is clicked
 ✔ returns to the add stop form when Cancel is clicked in edit mode
 ✔ calls createStop then addTagToStop for each tag when the add form is submitted
 ✔ calls updateStop then addTagToStop for each tag when the edit form is submitted
 ✔ calls removeTagFromStop for tags removed in the edit form
 ✔ renders List and Timeline tab buttons
 ✔ shows the stop list by default (List tab is active)
 ✔ switches to the timeline view when the Timeline tab is clicked
 ✔ returns to list view when the List tab is clicked after switching to Timeline
 ✔ renders an Export CSV button