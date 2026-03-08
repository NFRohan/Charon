import 'package:flutter/material.dart';

void main() {
  runApp(const CharonDriverApp());
}

class CharonDriverApp extends StatelessWidget {
  const CharonDriverApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Charon Driver',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFFB65C2E),
          brightness: Brightness.light,
        ),
        scaffoldBackgroundColor: const Color(0xFFF7EFE2),
        useMaterial3: true,
      ),
      home: const DriverHomePage(),
    );
  }
}

class DriverHomePage extends StatelessWidget {
  const DriverHomePage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Charon Driver'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text(
            'Driver Console',
            style: theme.textTheme.displaySmall,
          ),
          const SizedBox(height: 12),
          const _StatusTile(label: 'Bus attached', value: 'No bus selected'),
          const _StatusTile(label: 'Service label', value: 'Awaiting attach'),
          const _StatusTile(label: 'GPS', value: 'Permission required'),
          const _StatusTile(label: 'Network', value: 'Unknown'),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: () {},
            style: FilledButton.styleFrom(
              minimumSize: const Size.fromHeight(64),
            ),
            child: const Text('Attach Bus'),
          ),
          const SizedBox(height: 12),
          OutlinedButton(
            onPressed: () {},
            style: OutlinedButton.styleFrom(
              minimumSize: const Size.fromHeight(56),
            ),
            child: const Text('Start Journey'),
          ),
        ],
      ),
    );
  }
}

class _StatusTile extends StatelessWidget {
  const _StatusTile({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
      ),
      child: ListTile(
        title: Text(label),
        subtitle: Text(
          value,
          style: theme.textTheme.bodyLarge,
        ),
      ),
    );
  }
}
