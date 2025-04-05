- Clean up logging
- Add in timezone support
- Go through and have comments added. Make sure you understand the logic. This is the opportunity to learn. 
- Move to server
    - Put in docker
    - put behind SWAG

- Can this be a mobile app?
- Mobile App Options
    - Possible Approaches:
        1. React Native + Existing Backend
            - Reuse current Go backend as API
            - Cross-platform (iOS/Android)
            - Large ecosystem of UI components
        2. Flutter + Existing Backend
            - Single codebase for iOS/Android
            - Better performance than React Native
            - Material Design and Cupertino widgets
        3. Native Apps (Swift/Kotlin)
            - Best performance and UX
            - Full platform feature access
            - HealthKit/Google Fit integration

    - Key Features to Consider:
        - Offline Support
            - Local SQLite database
            - Background sync
        - Health Platform Integration
            - Apple HealthKit (iOS)
            - Google Fit (Android)
        - Mobile-Specific Features
            - Push notifications for reminders
            - Widgets for quick readings
            - Share with healthcare providers
            - PDF/CSV export
            - Biometric authentication
        - UI/UX Enhancements
            - Charts and graphs
            - Dark mode
            - Accessibility features
            - Quick input interface
            - Watch companion apps
