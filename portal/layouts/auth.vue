<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="5" lg="4">
        <v-card class="pa-6 pa-md-8" elevation="8" rounded="lg">
          <div class="text-center mb-6">
            <div class="text-h5 font-weight-bold text-primary">AdLite</div>
            <div class="text-caption text-medium-emphasis">DOOH Advertising Platform</div>
          </div>
          <slot />
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>
