import $RefParser from '@apidevtools/json-schema-ref-parser';
import yaml from 'js-yaml';

(async () => {
  try {
    const json = yaml.load('../apisrv/api/issuesv1.schema.yaml');
    const schema = await $RefParser.dereference(json);
    const examples: unknown[] = [];

    for (const ref of schema.anyOf) {
      if (typeof ref === 'boolean') continue;

      for (const ex of ref.examples || []) {
        examples.push(ex);
      }
    }

    const code = JSON.stringify(examples, null, 2);
    console.log(`import * as IssuesV1 from 'generated/issuesv1';

export const examples: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml[] = ${code};
`);
  } catch (err) {
    console.error(err);
  }
})();
