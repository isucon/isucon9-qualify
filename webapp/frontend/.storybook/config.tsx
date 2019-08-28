import {addDecorator, configure} from '@storybook/react';
import {MuiThemeProvider} from "@material-ui/core";
import {themeInstance} from "../src/theme";
import * as React from "react";

// automatically import all files ending in *.stories.tsx
const req = require.context('../src', true, /\.stories\.tsx$/);
function loadStories() {
  req.keys().forEach(req);
}

const BaseDecorator = storyFn => (
  <MuiThemeProvider theme={themeInstance}>
    {storyFn()}
  </MuiThemeProvider>
);
addDecorator(BaseDecorator);

configure(loadStories, module);
