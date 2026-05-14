export default {
  extends: ['stylelint-config-standard'],
  ignoreFiles: ['node_modules/**', 'dist/**'],
  overrides: [
    {
      files: ['**/*.vue'],
      customSyntax: 'postcss-html',
    },
  ],
  rules: {
    'at-rule-no-unknown': [
      true,
      {
        ignoreAtRules: ['theme', 'utility', 'variant', 'custom-variant', 'apply', 'layer', 'tailwind'],
      },
    ],
    'alpha-value-notation': null,
    'custom-property-empty-line-before': null,
    'import-notation': null,
    'no-descending-specificity': null,
    'selector-not-notation': null,
    'value-keyword-case': null,
  },
}
