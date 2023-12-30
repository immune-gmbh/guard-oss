import { describe, expect, test } from '@jest/globals';
import { convertToCamelRecursive, convertToSnakeRecursive } from 'utils/case';

describe('case utils', () => {

  test('handle cyclic structures', () => {
    const a = { hello: 'world' };
    const c = { helloWorld: 'TestValue', c: [a, a, a] };
    const b = { hello_world: 'another_test', a, c };
    a['b'] = b;

    convertToCamelRecursive(a);

    expect(b['helloWorld']).toEqual('another_test');
    expect(b.hello_world).toEqual(undefined);
    convertToSnakeRecursive(a);
    expect(c['hello_world']).toEqual('TestValue');
    expect(c.helloWorld).toEqual(undefined);
  });

  test('handle arrays', () => {
    const a = ['a', 'b', 'c'];

    convertToCamelRecursive(a);
    expect(a).toEqual(['a', 'b', 'c']);
  });
})
