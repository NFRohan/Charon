import 'package:flutter_test/flutter_test.dart';

import 'package:driver_app/main.dart';

void main() {
  testWidgets('renders driver scaffold shell', (WidgetTester tester) async {
    await tester.pumpWidget(const CharonDriverApp());

    expect(find.text('Charon Driver'), findsOneWidget);
    expect(find.text('Attach Bus'), findsOneWidget);
  });
}
