import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { ItemFooter } from '.';
import { Typography } from '@material-ui/core';
import { ReactElement } from 'react';

const stories = storiesOf('components/ItemFooter', module);

const getMockButton = (
  text: string,
  disabled: boolean,
  tooltip?: ReactElement,
) => ({
  onClick: (e: React.MouseEvent) => {},
  buttonText: text,
  disabled,
  tooltip,
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
  ))
  .add('with tooltip', () => (
    <ItemFooter
      {...mockProps}
      buttons={[
        getMockButton(
          'BUMP',
          false,
          <React.Fragment>
            <Typography variant="subtitle1">新機能！</Typography>
            <Typography variant="subtitle2">
              BUMPを使って商品をタイムラインの一番上に押し上げよう！'
            </Typography>
          </React.Fragment>,
        ),
      ]}
    />
  ));
