import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { SnackBar } from '.';

const stories = storiesOf('components/SnackBar', module);

stories.add('default', () => (
  <SnackBar
    open={true}
    message={'message'}
    handleClose={(e: React.MouseEvent) => {}}
  />
));
