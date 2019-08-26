import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { InternalServerErrorPage } from '.';
import { MemoryRouter } from 'react-router-dom';

const stories = storiesOf('pages/InternalServerErrorPage', module);

stories.add('default', () => (
  <MemoryRouter>
    <InternalServerErrorPage />
  </MemoryRouter>
));

stories.add('with message', () => (
  <MemoryRouter>
    <InternalServerErrorPage message={'something wrong...'} />
  </MemoryRouter>
));
