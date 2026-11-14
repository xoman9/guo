import 'package:flutter/material.dart';

class PageHeader extends StatelessWidget {
  const PageHeader({super.key, required this.title, this.actions = const [], this.titleWidget});
  final String title;
  final Widget? titleWidget;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.fromLTRB(20, 0, 8, 0),
    child: SizedBox(
      height: 64,
      child: Row(
        children: [
          Expanded(
            child: Semantics(
              header: true,
              child: titleWidget ?? Text(
                title,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontSize: 28,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),
          ...actions,
        ],
      ),
    ),
  );
}
