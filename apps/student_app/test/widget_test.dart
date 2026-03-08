import 'package:flutter_test/flutter_test.dart';

import 'package:student_app/main.dart';

void main() {
  testWidgets('renders student scaffold shell', (WidgetTester tester) async {
    await tester.pumpWidget(const CharonStudentApp());

    expect(find.text('Charon Student'), findsOneWidget);
    expect(find.text('Scan to Pay'), findsOneWidget);
  });
}
