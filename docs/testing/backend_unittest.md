cd backend && go test -json ./... | gotestdox
github.com/pkordes/rv-logbook/backend/internal/config:
 ✔ Load defaults (0.00s)
 ✔ Load missing required (0.00s)
 ✔ Load overrides (0.00s)

github.com/pkordes/rv-logbook/backend/internal/middleware:
 ✔ CORSHandler GET allowed origin (0.00s)
 ✔ CORSHandler GET disallowed origin (0.00s)
 ✔ CORSHandler OPTIONS preflight (0.00s)
 ✔ MaxBodySizeHandler content length exceeds limit returns 413 (0.00s)
 ✔ MaxBodySizeHandler small body passes through (0.00s)
 ✔ MaxBodySizeHandler streaming body exceeds limit returns 413 (0.00s)
 ✔ SecurityHeaders passes through status (0.00s)
 ✔ SecurityHeaders sets expected headers (0.00s)
 ✔ SlogLogger logs request fields (0.00s)

github.com/pkordes/rv-logbook/backend/internal/handler:
 ✔ AddTagToStop 201 (0.00s)
 ✔ AddTagToStop 404 stop not found (0.00s)
 ✔ AddTagToStop 422 validation error (0.00s)
 ✔ CreateStop 201 (0.00s)
 ✔ CreateStop 404 trip not found (0.00s)
 ✔ CreateStop 422 validation (0.00s)
 ✔ CreateTag 201 (0.00s)
 ✔ CreateTag 422 empty name (0.00s)
 ✔ CreateTrip 201 (0.00s)
 ✔ CreateTrip 422 validation error (0.00s)
 ✔ DeleteStop 204 (0.00s)
 ✔ DeleteStop 404 (0.00s)
 ✔ DeleteTag 204 (0.00s)
 ✔ DeleteTag 404 (0.00s)
 ✔ DeleteTrip 204 (0.00s)
 ✔ DeleteTrip 404 (0.00s)
 ✔ GetExport CSV empty result has header row (0.00s)
 ✔ GetExport CSV format param content type (0.00s)
 ✔ GetExport CSV one row has header and data row (0.00s)
 ✔ GetExport CSV tags joined with pipe (0.00s)
 ✔ GetExport JSON trip with no stops empty stop fields (0.00s)
 ✔ GetExport default JSON empty result (0.00s)
 ✔ GetExport format JSON explicit param (0.00s)
 ✔ GetExport service error returns 500 (0.00s)
 ✔ GetHealth returns 200 with OK status (0.00s)
 ✔ GetStop 200 (0.00s)
 ✔ GetStop 404 (0.00s)
 ✔ GetTrip 200 (0.00s)
 ✔ GetTrip 404 (0.00s)
 ✔ ListStops 200 (0.00s)
 ✔ ListStops 200 empty (0.00s)
 ✔ ListTags 200 (0.00s)
 ✔ ListTags 200 with prefix (0.00s)
 ✔ ListTagsByStop 200 (0.00s)
 ✔ ListTrips 200 (0.00s)
 ✔ ListTrips 200 empty (0.00s)
 ✔ PatchTag 200 (0.00s)
 ✔ PatchTag 404 (0.00s)
 ✔ PatchTag 422 empty name (0.00s)
 ✔ RemoveTagFromStop 204 (0.00s)
 ✔ RemoveTagFromStop 404 not linked (0.00s)
 ✔ UpdateStop 200 (0.00s)
 ✔ UpdateStop 404 (0.00s)
 ✔ UpdateTrip 200 (0.00s)
 ✔ UpdateTrip 404 (0.00s)

github.com/pkordes/rv-logbook/backend/internal/service:
 ✔ ExportService export multiple trips multiple stops (0.00s)
 ✔ ExportService export no trips (0.00s)
 ✔ ExportService export one trip one stop no tags (0.00s)
 ✔ ExportService export stop with tags (0.00s)
 ✔ ExportService export trip end date included (0.00s)
 ✔ ExportService export trip repo error (0.00s)
 ✔ ExportService export trip with no stops (0.00s)
 ✔ StopService add tag OK (0.00s)
 ✔ StopService add tag add to stop error (0.00s)
 ✔ StopService add tag empty name (0.00s)
 ✔ StopService add tag normalizes name (0.00s)
 ✔ StopService add tag upsert error (0.00s)
 ✔ StopService create OK (0.00s)
 ✔ StopService create departed before arrived (0.00s)
 ✔ StopService create name required (0.00s)
 ✔ StopService create repo error (0.00s)
 ✔ StopService create trip not found (0.00s)
 ✔ StopService delete OK (0.00s)
 ✔ StopService delete not found (0.00s)
 ✔ StopService get by ID OK (0.00s)
 ✔ StopService get by ID not found (0.00s)
 ✔ StopService list by trip ID OK (0.00s)
 ✔ StopService list by trip ID returns empty slice (0.00s)
 ✔ StopService list tags by stop OK (0.00s)
 ✔ StopService list tags by stop returns empty slice (0.00s)
 ✔ StopService remove tag from stop OK (0.00s)
 ✔ StopService remove tag from stop not found (0.00s)
 ✔ StopService update OK (0.00s)
 ✔ StopService update validation fails (0.00s)
 ✔ TagService delete OK (0.00s)
 ✔ TagService delete not found (0.00s)
 ✔ TagService list all (0.00s)
 ✔ TagService list prefix normalized (0.00s)
 ✔ TagService list returns empty slice (0.00s)
 ✔ TagService update name OK (0.00s)
 ✔ TagService update name empty name (0.00s)
 ✔ TagService update name not found (0.00s)
 ✔ TagService upsert by name OK (0.00s)
 ✔ TagService upsert by name collapses punctuation (0.00s)
 ✔ TagService upsert by name empty after normalization (0.00s)
 ✔ TagService upsert by name empty name (0.00s)
 ✔ TagService upsert by name normalizes case (0.00s)
 ✔ TripService create end date before start date (0.00s)
 ✔ TripService create end date equal to start date (0.00s)
 ✔ TripService create missing name (0.00s)
 ✔ TripService create nil end date (0.00s)
 ✔ TripService create repo error (0.00s)
 ✔ TripService create valid (0.00s)
 ✔ TripService delete OK (0.00s)
 ✔ TripService delete not found (0.00s)
 ✔ TripService get by ID found (0.00s)
 ✔ TripService get by ID not found (0.00s)
 ✔ TripService list (0.00s)
 ✔ TripService list empty (0.00s)
 ✔ TripService update end date before start date (0.00s)
 ✔ TripService update missing name (0.00s)
 ✔ TripService update valid (0.00s)

