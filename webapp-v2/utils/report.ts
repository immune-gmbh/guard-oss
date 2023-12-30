export function sortIPv4(a: string, b: string): number {
  const order = ['public', 'private', 'link-local', 'loopback'];
  const getOrder = (ip: string): string => {
    if (ip.startsWith('169.254.')) {
      return 'link-local';
    } else if (ip.startsWith('127.')) {
      return 'loopback';
    } else if (ip.startsWith('192.168.')) {
      return 'private';
    } else if (ip.startsWith('10.')) {
      return 'private';
    } else if (ip.startsWith('172.')) {
      const b2 = ip.slice(4, 3);
      const n2 = parseInt(b2);

      if (b2.endsWith('.') && n2 >= 16 && n2 <= 31) {
        return 'private';
      } else {
        return 'public';
      }
    } else {
      return 'public';
    }
  };
  const o1 = order.indexOf(getOrder(a));
  const o2 = order.indexOf(getOrder(b));

  return o1 - o2;
}
