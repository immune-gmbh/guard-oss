import { snake, camel } from 'kitsu-core';

type ConvFn = (name: string) => string;

// Internal Utility Functions
const traverseObj = (obj: unknown, conv: ConvFn, occ: Set<unknown>): void => {
  const traverseArr = (arr: Array<unknown>, conv: ConvFn, occ: Set<unknown>): void => {
    arr.forEach((v) => {
      if (v && !occ.has(v)) {
        occ.add(v);
        if (Array.isArray(v)) {
          traverseArr(v, conv, occ);
        } else if (typeof v === 'object') {
          traverseObj(v, conv, occ);
        }
      }
    });
  };

  Object.keys(obj).forEach((k) => {
    if (obj[k] && !occ.has(obj[k])) {
      occ.add(obj[k]);
      if (Array.isArray(obj[k])) {
        traverseArr(obj[k], conv, occ);
      } else if (typeof obj[k] === 'object') {
        traverseObj(obj[k], conv, occ);
      }
    }

    const sck = conv(k);
    if (sck !== k) {
      obj[sck] = obj[k];
      delete obj[k];
    }
  });
};

// Internal Utility Functions
export function convertToSnakeRecursive(o: unknown): unknown {
  if (!o || typeof o !== 'object') return o;

  const occured = new Set();

  if (Array.isArray(o)) {
    const ary = [...o];
    traverseObj(ary, snake, occured);
    return ary;
  } else {
    const obj = Object.assign({}, o);
    traverseObj(obj, snake, occured);
    return obj;
  }
}

export function convertToCamelRecursive(o: unknown): unknown {
  if (!o || typeof o !== 'object') return o;

  const occured = new Set();

  if (Array.isArray(o)) {
    const ary = [...o];
    traverseObj(ary, camel, occured);
    return ary;
  } else {
    const obj = Object.assign({}, o);
    traverseObj(obj, camel, occured);
    return obj;
  }
}
