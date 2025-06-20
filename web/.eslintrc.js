module.exports = {
  extends: [
    'react-app',
    'react-app/jest'
  ],
  ignorePatterns: [
    'src/generated/**/*'
  ],
  rules: {
    // Disable some rules that are too strict for development
    'no-unused-vars': 'warn',
    'react-hooks/exhaustive-deps': 'warn'
  },
  settings: {
    react: {
      version: 'detect'
    }
  }
};