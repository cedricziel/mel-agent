import js from '@eslint/js';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import globals from 'globals';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';

export default [
  {
    ignores: ['dist', 'node_modules', 'coverage', 'cypress/videos', 'cypress/screenshots', 'src/generated/**']
  },
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: {
        ...globals.browser,
        ...globals.es2020,
        ...globals.node
      },
      parserOptions: {
        ecmaVersion: 'latest',
        ecmaFeatures: {
          jsx: true
        },
        sourceType: 'module'
      }
    },
    settings: {
      react: {
        // Pinned instead of 'detect': eslint-plugin-react's version
        // auto-detection uses a context API removed in ESLint 10 and crashes.
        // Revert to 'detect' (and unpin) once eslint-plugin-react ships a
        // release whose peer range includes eslint 10.
        version: '19.1'
      }
    },
    plugins: {
      react,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh
    },
    rules: {
      ...js.configs.recommended.rules,
      ...react.configs.recommended.rules,
      ...react.configs['jsx-runtime'].rules,
      ...reactHooks.configs.recommended.rules,

      // Custom rules
      'no-unused-vars': 'warn',
      // New-in-ESLint-10 / react-hooks-7 rules downgraded to warnings for
      // pre-existing violations; burn these down and restore to errors.
      'preserve-caught-error': 'warn',
      'react-hooks/exhaustive-deps': 'warn',
      'react-hooks/set-state-in-effect': 'warn',
      'react/prop-types': 'off', // Turn off prop-types as we might be using TypeScript or don't need it
      'react/react-in-jsx-scope': 'off', // Not needed with React 17+
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true }
      ]
    }
  },
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true
        }
      }
    },
    plugins: {
      '@typescript-eslint': tseslint,
      react,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh
    },
    rules: {
      ...js.configs.recommended.rules,
      ...tseslint.configs.recommended.rules,
      ...react.configs.recommended.rules,
      ...react.configs['jsx-runtime'].rules,
      ...reactHooks.configs.recommended.rules,
      
      // TypeScript specific rules
      '@typescript-eslint/no-unused-vars': 'warn',
      'no-unused-vars': 'off', // Turn off base rule as it conflicts with TypeScript rule
      // See the js block above: downgraded pending burn-down.
      'preserve-caught-error': 'warn',
      'react-hooks/set-state-in-effect': 'warn',
      'react/prop-types': 'off', // Not needed with TypeScript
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true }
      ]
    }
  },
  {
    files: ['**/*.test.{js,jsx,ts,tsx}', '**/__tests__/**/*.{js,jsx,ts,tsx}'],
    languageOptions: {
      globals: {
        ...globals.jest,
        ...globals.browser
      }
    },
    rules: {
      // Test-specific rules can be added here
    }
  }
];
