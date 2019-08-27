import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { NotFoundPage } from '.';
import { MemoryRouter } from 'react-router-dom';

const stories = storiesOf('pages/NotFoundPage', module);

stories.add('default', () => (
  <MemoryRouter>
    <NotFoundPage />
  </MemoryRouter>
));

stories.add('with message', () => (
  <MemoryRouter>
    <NotFoundPage message={'not found!'} />
  </MemoryRouter>
));
