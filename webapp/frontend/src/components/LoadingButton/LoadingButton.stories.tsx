import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { LoadingButton } from '.';

const stories = storiesOf('components/LoadingButton', module);

const mockProps = {
  onClick: (e: React.MouseEvent) => {},
};

stories
  .add('default', () => (
    <LoadingButton {...mockProps} buttonName="default" loading={false} />
  ))
  .add('loading', () => (
    <LoadingButton {...mockProps} buttonName="loading" loading={true} />
  ));
