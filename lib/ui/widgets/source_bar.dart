import 'package:flutter/material.dart';

import '../app_scope.dart';
import 'bar_button.dart';
import 'page_header.dart';
import 'search_field.dart';

class SourceBar extends StatelessWidget {
  const SourceBar({
    super.key,
    required this.source,
    required this.onSwitchSource,
    required this.onOpenSearch,
    required this.onSearchSubmitted,
    this.searchController,
    this.televisionLeading,
    this.leading = const [],
    this.actions = const [],
  });
  final List<Widget> leading;
  final String source;
  final void Function(BuildContext anchor) onSwitchSource;
  final VoidCallback onOpenSearch;
  final ValueChanged<String> onSearchSubmitted;
  final TextEditingController? searchController;
  final Widget? televisionLeading;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) {
    final form = AppScope.formFactorOf(context);
    final hint = '在 $source 中搜索';
    final sourceSwitcher = Builder(
      builder: (anchor) => TextButton(
        onPressed: () => onSwitchSource(anchor),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Flexible(
              child: Text(
                source,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontWeight: FontWeight.w700),
              ),
            ),
            const SizedBox(width: 2),
            const Icon(Icons.arrow_drop_down_rounded, size: 21),
          ],
        ),
      ),
    );
    final search = BarButton(
      icon: Icons.search_rounded,
      tooltip: hint,
      onPressed: onOpenSearch,
    );
    return switch (form) {
      FormFactor.phone => PageHeader(
        title: source,
        titleWidget: Align(alignment: Alignment.centerLeft, child: sourceSwitcher),
        actions: [...leading, search, ...actions],
      ),
      FormFactor.pad || FormFactor.window => PageHeader(
        title: source,
        titleWidget: Align(alignment: Alignment.centerLeft, child: sourceSwitcher),
        actions: [
          ...leading,
          const SizedBox(width: 4),
          SizedBox(
            width: MediaQuery.sizeOf(context).width >= 900 ? 320 : 240,
            child: SearchField(
              hint: hint,
              controller: searchController,
              onSubmitted: onSearchSubmitted,
            ),
          ),
          const SizedBox(width: 4),
          ...actions,
        ],
      ),
      FormFactor.television => SizedBox(
        height: 48,
        child: Row(
          spacing: 10,
          children: [
            ?televisionLeading,
            sourceSwitcher,
            const Spacer(),
            search,
            ...actions,
          ],
        ),
      ),
    };
  }
}
