import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { ItemFooter } from '.';

const stories = storiesOf('components/ItemFooter', module);

const getMockButton = (text: string, disabled: boolean) => ({
  onClick: (e: React.MouseEvent) => {},
  buttonText: text,
  disabled,
});

const mockProps = {
  price: 1000,
  buttons: [getMockButton('購入', false)],
};

stories
  .add('single button', () => <ItemFooter {...mockProps} />)
  .add('multiple button', () => (
    <ItemFooter
      {...mockProps}
      buttons={[getMockButton('購入', false), getMockButton('Bump', false)]}
    />
  ))
  .add('disabled button', () => (
    <ItemFooter {...mockProps} buttons={[getMockButton('購入', true)]} />
  ));
