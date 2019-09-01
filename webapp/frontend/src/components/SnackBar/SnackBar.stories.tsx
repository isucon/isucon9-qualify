import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { SnackBar } from '.';

const stories = storiesOf('components/SnackBar', module);

stories
  .add('default', () => (
    <SnackBar
      open={true}
      variant="success"
      message={'message'}
      handleClose={(e: React.MouseEvent) => {}}
    />
  ))
  .add('error', () => (
    <SnackBar
      open={true}
      variant="error"
      message={'error'}
      handleClose={(e: React.MouseEvent) => {}}
    />
  ));
