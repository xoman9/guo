import 'package:duanju_app/ui/player/swipe_controls.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('portrait swipe zones control brightness, episodes, and volume', () {
    expect(portraitSwipeZone(10, 400), PortraitSwipeZone.brightness);
    expect(portraitSwipeZone(119, 400), PortraitSwipeZone.brightness);
    expect(portraitSwipeZone(120, 400), PortraitSwipeZone.episodes);
    expect(portraitSwipeZone(200, 400), PortraitSwipeZone.episodes);
    expect(portraitSwipeZone(280, 400), PortraitSwipeZone.episodes);
    expect(portraitSwipeZone(281, 400), PortraitSwipeZone.volume);
    expect(portraitSwipeZone(390, 400), PortraitSwipeZone.volume);
  });

  test('non-positive widths cannot activate either system level', () {
    expect(portraitSwipeZone(0, 0), PortraitSwipeZone.episodes);
    expect(portraitSwipeZone(-4, -10), PortraitSwipeZone.episodes);
  });
}
