import 'package:duanju_app/danmaku_models.dart';
import 'package:duanju_app/ui/player/danmaku_overlay.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('common bracketed Hongguo reactions render as emojis', () {
    expect(
      renderDanmakuReactions('好看[拍手][捂脸][笑哭]'),
      '好看👏🤦😂',
    );
    expect(
      renderDanmakuReactions('[捂臉][愛心][点赞][哭]'),
      '🤦❤️👍😭',
    );
    expect(renderDanmakuReactions('原样[未知表情]文字'), '原样[未知表情]文字');
    expect(renderDanmakuReactions('已经有😂 [拍手]'), '已经有😂 👏');
    expect(renderDanmakuReactions('[吃瓜][酷][送花]'), '🍉😎💐');
    expect(renderDanmakuReactions('[奸笑][愛慕][快哭了][怒]'), '😏😍🥺😡');
    expect(renderDanmakuReactions('[伤心][DOGE]'), '😢🐶');
    expect(renderDanmakuReactions('[傷心][doge]'), '😢🐶');
  });

  test('flight uses emoji display while preserving source identity and time', () {
    const source = DanmakuItem(id: 'comment-1', text: '精彩[拍手]', timeMs: 2000);
    final measured = <String>[];
    final flights = planDanmaku(
      [source],
      width: 600,
      rows: 2,
      measure: (text) {
        measured.add(text);
        return 64;
      },
    );
    expect(flights, hasLength(1));
    expect(flights.single.text, '精彩👏');
    expect(flights.single.item, same(source));
    expect(flights.single.item.text, '精彩[拍手]');
    expect(flights.single.item.timeMs, 2000);
    expect(measured, ['精彩👏']);
  });
}
