package core

import "testing"

// A successful compile must not silently drop custom channels from the
// upstream dynamic Registry. This checks registrations without accessing sites.
func TestReleaseReplayCustomChannelsRegistered(t *testing.T) {
 d := &Downloader{cfg: DefaultConfig()}
 ids := []string{sourceLujj31, sourceLiuguo, sourceXiaoguo}
 for _, spec := range apk391Sources { ids = append(ids, spec.ID) }
 seen := map[string]bool{}
 for _, id := range ids {
  if seen[id] { t.Fatalf("duplicate custom source %q", id) }
  seen[id] = true
  descriptor, ok := registeredDescriptor(id)
  if !ok || descriptor.ID != id { t.Fatalf("source %q not registered", id) }
  provider, ok := d.provider(id)
  if !ok || provider == nil || provider.Descriptor().ID != id {
   t.Fatalf("source %q has no functional provider", id)
  }
  if !nativeSourceAvailable(id) { t.Fatalf("source %q unavailable in unrestricted build", id) }
 }
 if len(ids) != 13 { t.Fatalf("expected 13 replayed channels, got %d", len(ids)) }
}
