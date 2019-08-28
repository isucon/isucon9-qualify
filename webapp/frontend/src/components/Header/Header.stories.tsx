import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { Header } from '.';

const stories = storiesOf('components/Header', module);

const mockProps = {
  isLoggedIn: false,
  ownUserId: 1111,
  goToTopPage: () => {},
  goToUserPage: (userId: number) => {},
  goToSettingPage: () => {},
};

stories
  .add('non sign in', () => <Header {...mockProps} />)
  .add('signed in', () => <Header {...mockProps} isLoggedIn={true} />);
