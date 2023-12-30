import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';
import { setConfig } from 'next/config';
import 'whatwg-fetch';
import 'regenerator-runtime/runtime';

import config from './next.config';

setConfig(config);
