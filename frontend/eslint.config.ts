import js from '@eslint/js';
import { configureVueProject, defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript';
import prettierConfig from 'eslint-config-prettier/flat';
import perfectionist from 'eslint-plugin-perfectionist';
import unusedImports from 'eslint-plugin-unused-imports';
import vue from 'eslint-plugin-vue';
import { globalIgnores } from 'eslint/config';
import globals from 'globals';

configureVueProject({
    tsSyntaxInTemplates: true,
    scriptLangs: ['ts'],
    allowComponentTypeUnsafety: false,
    rootDir: import.meta.dirname,
});

export default defineConfigWithVueTs(
    js.configs.recommended,
    vue.configs['flat/recommended'],
    vueTsConfigs.strictTypeChecked,
    vueTsConfigs.stylisticTypeChecked,

    globalIgnores(['node_modules', 'dist', 'bindings', '.vite', '**/*.d.ts', '*.config.js', '*.config.ts']),

    {
        name: 'custom',
        files: [
            'src/**/*.ts',
            'src/**/*.mts',
            'src/**/*.cts',
            'src/**/*.js',
            'src/**/*.mjs',
            'src/**/*.cjs',
            'src/**/*.vue',
        ],
        plugins: { perfectionist, 'unused-imports': unusedImports },
        linterOptions: { reportUnusedDisableDirectives: 'error', reportUnusedInlineConfigs: 'error' },
        languageOptions: {
            ecmaVersion: 'latest',
            sourceType: 'module',
            globals: globals.browser,
            parserOptions: {
                projectService: true,
                tsconfigRootDir: import.meta.dirname,
            },
        },
        rules: {
            // ESLint core rules
            'block-scoped-var': 'error',
            'class-methods-use-this': 'off',
            curly: 'error',
            'default-case-last': 'error',
            'default-param-last': 'off',
            'dot-notation': 'error',
            eqeqeq: 'error',
            'init-declarations': 'off',
            'no-alert': 'error',
            // A failure that only costs a feature is printed and swallowed here, see the Go side of it.
            'no-console': ['error', { allow: ['warn', 'error'] }],
            'no-extra-label': 'error',
            'no-implied-eval': 'error',
            'no-invalid-this': 'error',
            'no-loop-func': 'off',
            'no-param-reassign': 'error',
            'no-restricted-syntax': ['error', 'SequenceExpression'],
            'no-shadow': 'off',
            'no-template-curly-in-string': 'error',
            'no-unassigned-vars': 'error',
            'no-unneeded-ternary': 'error',
            'no-unused-expressions': 'error',
            'no-useless-assignment': 'error',
            'no-useless-concat': 'error',
            'no-useless-constructor': 'error',
            'no-useless-return': 'error',
            // Helpers are declared below the function that uses them throughout this project.
            'no-use-before-define': ['error', { functions: false }],
            'no-var': 'error',
            'prefer-template': 'error',

            // TypeScript ESLint rules
            '@typescript-eslint/class-methods-use-this': 'error',
            '@typescript-eslint/consistent-generic-constructors': ['error', 'type-annotation'],
            '@typescript-eslint/consistent-type-assertions': [
                'error',
                { assertionStyle: 'as', objectLiteralTypeAssertions: 'never', arrayLiteralTypeAssertions: 'never' },
            ],
            '@typescript-eslint/consistent-type-exports': 'error',
            '@typescript-eslint/consistent-type-imports': 'error',
            '@typescript-eslint/default-param-last': 'error',
            '@typescript-eslint/explicit-function-return-type': 'off',
            '@typescript-eslint/explicit-member-accessibility': 'error',
            '@typescript-eslint/explicit-module-boundary-types': 'off',
            '@typescript-eslint/init-declarations': 'error',
            '@typescript-eslint/method-signature-style': 'error',
            '@typescript-eslint/naming-convention': [
                'error',
                {
                    selector: 'default',
                    format: ['camelCase'],
                },
                {
                    selector: 'import',
                    format: ['camelCase', 'PascalCase'],
                },
                {
                    selector: 'variable',
                    format: ['camelCase', 'UPPER_CASE'],
                },
                {
                    selector: 'typeLike',
                    format: ['PascalCase'],
                },
                {
                    // `update:modelValue` and its kind are prescribed by Vue and have to be quoted.
                    selector: ['objectLiteralProperty', 'typeProperty'],
                    modifiers: ['requiresQuotes'],
                    format: null,
                },
            ],
            '@typescript-eslint/no-import-type-side-effects': 'error',
            '@typescript-eslint/no-loop-func': 'error',
            // The actions the components are handed are typed `() => void` although they are async.
            // They never reject, because `data/request.ts` reports a failed call instead of throwing,
            // so a promise that is not awaited is the regular case wherever one is passed on.
            '@typescript-eslint/no-misused-promises': [
                'error',
                { checksVoidReturn: { arguments: false, attributes: false, properties: false } },
            ],
            '@typescript-eslint/no-non-null-assertion': 'off',
            '@typescript-eslint/no-shadow': 'error',
            '@typescript-eslint/no-unnecessary-parameter-property-assignment': 'error',
            '@typescript-eslint/no-unnecessary-qualifier': 'error',
            '@typescript-eslint/prefer-enum-initializers': 'error',
            '@typescript-eslint/prefer-readonly': 'error',
            '@typescript-eslint/require-array-sort-compare': 'error',
            '@typescript-eslint/switch-exhaustiveness-check': [
                'error',
                {
                    allowDefaultCaseForExhaustiveSwitch: true,
                    requireDefaultForNonUnion: true,
                    considerDefaultExhaustiveForUnions: true,
                    defaultCaseCommentPattern: '/^no default$/iu',
                },
            ],

            // Import rules
            // These replace `prettier-plugin-organize-imports`, which never reached the script
            // blocks of the SFCs, so every `.vue` file drifted out of order. Both plugins sorting at
            // once would fight over every file, so the prettier one is gone.
            'perfectionist/sort-imports': [
                'error',
                {
                    // Everything that is not relative first, relative last, alphabetically within
                    // both. That is the order the project already had.
                    groups: [
                        ['builtin', 'external', 'internal', 'tsconfig-path', 'subpath', 'unknown'],
                        ['parent', 'sibling', 'index', 'style'],
                    ],
                    ignoreCase: true,
                    newlinesBetween: 'ignore',
                    // Reads the aliases out of `tsconfig.json`, so `@lib/cn` is told apart from
                    // `@lucide/vue` without listing them a third time.
                    tsconfig: { rootDir: '.' },
                },
            ],
            // Values before types, which is what the majority of the files already do.
            'perfectionist/sort-named-imports': [
                'error',
                { groups: ['value-import', 'type-import'], ignoreCase: true },
            ],
            // An import that is no longer used is deleted rather than only reported, which is the
            // one thing the prettier plugin did that sorting alone does not.
            '@typescript-eslint/no-unused-vars': 'off',
            'unused-imports/no-unused-imports': 'error',
            'unused-imports/no-unused-vars': [
                'error',
                { args: 'after-used', argsIgnorePattern: '^_', vars: 'all', varsIgnorePattern: '^_' },
            ],

            // Vue ESLint rules
            // Prettier sorts the attributes through `prettier-plugin-organize-attributes`, and its
            // order is not the one this rule wants.
            'vue/attributes-order': 'off',
            'vue/block-lang': [
                'error',
                {
                    script: {
                        lang: 'ts',
                    },
                },
            ],
            'vue/block-order': [
                'error',
                {
                    order: ['script', 'template', 'style'],
                },
            ],
            'vue/comment-directive': [
                'error',
                {
                    reportUnusedDisableDirectives: true,
                },
            ],
            'vue/component-api-style': ['error', ['script-setup', 'composition']],
            'vue/component-name-in-template-casing': ['error', 'PascalCase', { registeredComponentsOnly: true }],
            'vue/component-options-name-casing': ['error', 'PascalCase'],
            'vue/custom-event-name-casing': ['error', 'camelCase'],
            'vue/define-emits-declaration': ['error', 'type-literal'],
            'vue/define-macros-order': ['error'],
            'vue/define-props-declaration': ['error', 'type-based'],
            'vue/enforce-style-attribute': ['error'],
            'vue/html-button-has-type': ['error'],
            'vue/html-indent': ['error', 4],
            'vue/multi-word-component-names': 'off',
            'vue/next-tick-style': ['error', 'promise'],
            'vue/no-setup-props-reactivity-loss': ['error'],
            'vue/no-undef-components': ['error'],
            'vue/no-unused-properties': ['error', { groups: ['props', 'data', 'computed', 'methods', 'setup'] }],
            'vue/no-unused-refs': ['error'],
            'vue/no-useless-mustaches': ['error'],
            'vue/padding-line-between-blocks': ['error', 'always'],
            'vue/prefer-define-options': ['error'],
            'vue/prefer-separate-static-class': ['error'],
            'vue/prefer-true-attribute-shorthand': ['error'],
            'vue/require-typed-object-prop': ['error'],
            'vue/require-typed-ref': ['error'],
            'vue/slot-name-casing': ['error', 'camelCase'],
            'vue/v-for-delimiter-style': ['error', 'in'],
        },
    },
    {
        files: ['src/**/*.ts', 'src/**/*.mts', 'src/**/*.cts', 'src/**/*.vue'],
        rules: {
            '@typescript-eslint/explicit-function-return-type': 'error',
            '@typescript-eslint/explicit-module-boundary-types': 'error',
        },
    },
    prettierConfig,
);
